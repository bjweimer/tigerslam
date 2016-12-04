// Package colissionavoidance subscribes to LIDAR readings, inspects them and
// if something is within a "prohibited" part of the measured area, it sends
// off a "stop" signal. When the object is removed, a "start" signal is sent.
package collisionavoidance

import (
	"log"
	"math"

	"robot/config"
	"robot/logging"
	"robot/sensors/lidar"
	"robot/sensors/sensor"
)

var logger *log.Logger

func init() {
	logger = logging.New()
}

type CollisionDetector struct {
	angle  float64
	radius float64

	minIndex int
	maxIndex int

	// Internal flag
	isStopped bool

	// Channel for incoming LIDAR readings
	lidarChan chan sensor.SensorReading

	// Internal channel for stopping the go routine execution
	stopRoutineChan chan bool

	// Channels for outgoing STOP and RESUME signals
	StopChan   chan bool
	ResumeChan chan bool
}

// Create a CollisionDetector with angle and radius specified. The angle is in
// radians, the radius is in meters.
func MakeCollisionDetector(angle, radius float64) *CollisionDetector {
	c := &CollisionDetector{
		angle:  angle,
		radius: radius,
	}

	c.StopChan = make(chan bool)
	c.ResumeChan = make(chan bool)

	return c
}

// Make collision detector from default parameters
func MakeDefaultCollisionDetector() *CollisionDetector {
	return MakeCollisionDetector(config.COLLISION_DETECTION_ANGLE,
		config.COLLISION_DETECTION_RADIUS)
}

func (c *CollisionDetector) Start() {
	// Subscribe to LIDAR readings
	c.lidarChan = lidar.LidarSensor.Subscribe()

	nPerDeg := float64(lidar.LidarSensor.Distances) / lidar.LidarSensor.RadialSpan

	nPerRad := nPerDeg * 180 / math.Pi

	c.minIndex = lidar.LidarSensor.Distances/2 - int(nPerRad*c.angle/2)
	c.maxIndex = lidar.LidarSensor.Distances/2 + int(nPerRad*c.angle/2)

	c.stopRoutineChan = make(chan bool)

	// Run the routine until Stop() is called (return)
	go c.checkRoutine()

	logger.Println("Collision avoidance started")
}

func (c *CollisionDetector) Stop() {
	// Unsubscribe
	lidar.LidarSensor.Unsubscribe(c.lidarChan)

	// Stop routine
	c.stopRoutineChan <- true
}

func (c *CollisionDetector) Reset() {
	c.isStopped = false
}

func (c *CollisionDetector) checkRoutine() {

	var sensorReading sensor.SensorReading
	for {
		// Wait for either a LIDAR reading, or a signal for stopping the
		// execution.
		select {
		case sensorReading = <-c.lidarChan:
			// nothing
		case <-c.stopRoutineChan:
			return
		}

		// Convert to LIDAR reading
		lidarReading := sensorReading.(*lidar.LidarReading)

		if c.isStopped {
			// We're in the stopped state, and should send a RESUME signal if
			// the area is now non-occupied.
			if !c.areaOccupied(lidarReading) {
				c.isStopped = false

				c.ResumeChan <- true
				// select {
				// case c.ResumeChan <- true:
				// default:
				// }
				logger.Println("Obstacle gone")
			}
		} else {
			// We're not in the stopped state, and should send a STOP signal if
			// the area is occupied.
			if c.areaOccupied(lidarReading) {
				c.isStopped = true

				c.StopChan <- true
				// select {
				// case c.StopChan <- true:
				// default:
				// }
				logger.Println("!Detected obstacle!")
			}
		}
	}
}

// Check if the area is occupied (one or more laser beams are stopping)
func (c *CollisionDetector) areaOccupied(lr *lidar.LidarReading) bool {
	// Check if there's anything within that area
	for i := c.minIndex; i < c.maxIndex; i++ {
		if lr.Distances[i] > 10 && lr.Distances[i] < c.radius*1000 {
			return true
		}
	}

	return false
}
