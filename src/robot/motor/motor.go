package motor

import (
	"log"
	"sync"
	"time"

	"hectormapping/map/gridmap"

	"robot/collisionavoidance"
	"robot/fsm"
	"robot/logging"
	"robot/model"
	// "robot/motor/driver"
	"robot/motor/peasing"
	"robot/pathfollowing/lookahead"
	"robot/pathplanning"
	"robot/pathplanning/astar"
	// "robot/pathplanning/hybridastar"
	"robot/pathplanning/path"
	"robot/slam"
)

// States
const (
	MANUAL        = "MANUAL"
	PATHFOLLOWING = "PATHFOLLOWING"
)

var logger *log.Logger

func init() {
	logger = logging.New()
}

// The motor driver interface abstracts the motor driver. Structs implementing
// the motor driver interface should be thought of as a motor driver, and
// are either the motor driver itself or some other structure using the motor
// driver internally. This allows for implementing different modes of control
// for the motor, e.g. speed easing (not setting the motor speed directly, but
// smoothly), simulators etc.
type MotorDriver interface {
	// Make the motor ready for speed commands
	Connect() error

	// Disconnect the motor
	Disconnect() error

	// Determine if the motor is connected or not
	IsConnected() bool

	// Set the left and right speeds, in an interval [-1, 1] for each. The
	// interval represents the full allowable range of the motors, backwards
	// for negative values, forwards for positive values. Using values outside
	// the range may raise errors.
	SetSpeeds(left, right float64) error
}

// The Motor Controller provides an interface for the rest of the program to
// control the motors.
type MotorController struct {

	// State machine
	fsm.FSM

	// Used to set speeds of the motors. Might be a layer over the driver,
	// responsible for easing, so speeds set should be thaught of as references
	// for a controller and need not be smoothed.
	motor MotorDriver

	// Robot model for the path planner
	robot *model.DifferentialWheeledRobot

	// Pathplanner is responsible for calculating paths in a map.
	pathPlanner pathplanning.PathPlanner

	// Goal position for path planning in real world coordinates.
	goal [3]float64

	// Current planned path.
	path     *path.Path
	pathLock *sync.Mutex
	occMap   gridmap.OccGridMap
}

func MakeMotorController(robot *model.DifferentialWheeledRobot) *MotorController {
	mc := &MotorController{
		// motor: driver.MakeDefaultMotor(),
		motor: peasing.MakeDefaultPEasingMotorDriver(),
		robot: robot,
	}

	mc.pathLock = new(sync.Mutex)

	// Start off in manual state
	mc.SetState(MANUAL)

	return mc
}

func (m *MotorController) Disconnect() error {
	return m.motor.Disconnect()
}

// Set the speed of the motors to a value [-1, 1] where -1 is max reverse and
// 1 is max forward.
func (m *MotorController) ManualSpeeds(left, right float64) error {

	// Setting the state to manual can be done at any time, cancels any
	// path following.
	m.SetState(MANUAL)

	// Set the speeds and return any error
	return m.motor.SetSpeeds(left, right)
}

// Set goal and plan an initial path to this goal
func (m *MotorController) PlanPath(occMap gridmap.OccGridMap, currentLocation, goal [3]float64) error {

	var err error

	// If already following path, terminate it
	if m.GetState() == PATHFOLLOWING {
		m.StopPathFollowing()
	}

	m.goal = goal
	m.occMap = occMap

	p, err := m.planPathDirectly(currentLocation, m.goal)
	if err != nil {
		return err
	}

	m.pathLock.Lock()
	m.path = p
	m.pathLock.Unlock()

	return nil
}

// Plan a path from a location to another
func (m *MotorController) planPathDirectly(from, to [3]float64) (*path.Path, error) {

	pathPlanner := astar.MakeAstarPlanner(m.occMap, m.robot)
	return pathPlanner.PlanPath(from, to)

}

// Follow a path. Takes a SLAM algorithm as input, from which the continuously
// updated position is drawn.
func (m *MotorController) FollowPath(slamAlg slam.Slam) {

	logger.Println("Starting path following")

	if m.GetState() == PATHFOLLOWING {
		return
	}

	m.SetState(PATHFOLLOWING)

	// Set up collision avoidance
	collisionDetector := collisionavoidance.MakeDefaultCollisionDetector()
	collisionDetector.Start()

	go func() {
		defer collisionDetector.Stop()
		for {

			// Follow the current path
			outcome := m.followSubPath(collisionDetector, slamAlg)

			// Wait for the sub path following to finish
			successful := <-outcome

			if successful {
				logger.Println("Successfully arrived at goal.")
				m.SetState(MANUAL)
				return
			} else {

				// We did not arrive at goal as planned. Check if we have
				// switched to manual mode.
				if m.GetState() != PATHFOLLOWING || m.path == nil || slamAlg == nil {
					return
				} else {

					logger.Println("Starting backing")
					m.motor.SetSpeeds(-0.9, -0.9)

					// We need to wait with the replanning until the robot is
					// done backing, to get a good start coordinate.
					wg := new(sync.WaitGroup)
					wg.Add(1)
					time.AfterFunc(2*time.Second, func() {
						logger.Println("Stopping backing")
						m.motor.SetSpeeds(0, 0)
						wg.Done()
					})

					// Wait for backing.
					wg.Wait()

					// Plan new path to real goal
					currPos := slamAlg.GetPosition()
					replannedPath, err := m.planPathDirectly([3]float64{currPos.X, currPos.Y, currPos.Theta}, m.goal)
					if err != nil {
						logger.Println("Unable to find alternative route to goal.")
						m.SetState(MANUAL)
						return
					}

					// Set the new path and continue the journey
					m.pathLock.Lock()
					m.path = replannedPath
					m.pathLock.Unlock()

				}

			}
		}
	}()

}

// Follow a path stricly, return a channel which gives off a boolen value
// indicating if the path following was successful or not (i.e. if the robot)
// reached the end of the path. The collisiondetector is assumed to be
// started. If collisiondetector is nil, then no collision detection is
// performed (can be used to run small segments without collision detection).
func (m *MotorController) followSubPath(collisionDetector *collisionavoidance.CollisionDetector, slamAlg slam.Slam) chan bool {

	logger.Println("Starting following of subpath")
	successful := make(chan bool)

	follower := lookahead.MakeLookahead()
	follower.SetPath(m.path)

	interval := 100 * time.Millisecond

	// Loop this until finished or aborted
	go func() {

		// Print a message when stopping
		defer logger.Println("Stopped following of subpath")

		// The rate at which the speeds are updated
		ticker := time.NewTicker(interval)

		for {

			// Handle collision detection
			if collisionDetector != nil {
				select {
				case <-collisionDetector.StopChan:

					// Detected an obstacle, so stop the motors, wait for resume
					m.motor.SetSpeeds(0, 0)
					// logger.Println("Detected obstacle: stopping.")
					timer := time.NewTimer(5 * time.Second)
					select {
					case <-collisionDetector.ResumeChan:
						// continue
						// logger.Println("Resuming: obstacle gone.")
					case <-timer.C:
						collisionDetector.Reset()
						logger.Println("Aborting path")
						successful <- false
						return
					}
				default:
					// noop
				}
			}

			// Check if we're still in pathfollowing state
			if m.GetState() != PATHFOLLOWING || m.path == nil || slamAlg == nil {
				successful <- false
				return
			}

			// Lock the path, so it can't be changed or deleted
			m.pathLock.Lock()

			// Get the speed update
			pos := slamAlg.GetPosition()
			//logger.Printf("followSubPath: Current Position is X = %.3v Y = %.3v Theta = %.3v\n", pos.X, pos.Y, pos.Theta)
			v_l, v_r, finished := follower.SpeedUpdate([3]float64{pos.X, pos.Y, pos.Theta})

			// If we're finished
			if finished {
				m.motor.SetSpeeds(0, 0)
				m.pathLock.Unlock()
				successful <- true
				return
			}

			err := m.motor.SetSpeeds(v_l, v_r)
			if err != nil {
				logger.Println(err)
			}

			m.pathLock.Unlock()

			// Wait for tick
			<-ticker.C

		}

	}()

	return successful
}

// Stop path following
func (m *MotorController) StopPathFollowing() {
	// Setting state to manual will suffice, the
	m.SetState(MANUAL)
}

// Returns the current path (can be nil)
func (m *MotorController) GetPath() *path.Path {
	m.pathLock.Lock()
	defer m.pathLock.Unlock()
	return m.path
}

// Deletes the path -- is safe to run even when path following (any path
// following will be aborted).
func (m *MotorController) DeletePath() {

	m.pathLock.Lock()
	defer m.pathLock.Unlock()

	m.path = nil
}
