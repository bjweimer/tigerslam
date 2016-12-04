// Package slam provides SLAM imlementations.
//
// The package includes several SLAM implementations which implements a common
// interface for use with the rest of the program.
package slam

import (
	"errors"
	"image"
	"runtime"

	"hectormapping/map/maprep"

	"robot/fsm"
	"robot/mapstorage"
	"robot/model"
	"robot/slam/hector"
	// "robot/slam/tinyslam"
)

const (
	OFF     = "OFF"
	STOPPED = "STOPPED"
	RUNNING = "RUNNING"
)

type Slam interface {
	Start()
	Stop()
	GetPosition() model.Position
	// GetPositionHistory() []model.Position
	GetMapImage() (image.Image, error)
	GetMapTile(zoomLevel uint, tileX, tileY int) (image.Image, error)
	GetTypeName() string
	GetMapSizeMeters() float64
	GetMapSize() int
	GetOffsetX() float64
	GetOffsetY() float64
	GetMapRepresentation() maprep.MapRepresentation
}

type SlamController struct {
	fsm.FSM
	slam Slam
}

func MakeSlamController() *SlamController {
	return &SlamController{
		FSM: *fsm.MakeFSM(OFF),
	}
}

// Set up a slam algorithm, put the controller in STOPPED state
func (sc *SlamController) InitializeSlam(algorithm string, robot model.Robot) error {
	if sc.slam != nil {
		return errors.New("SLAM already initialized.")
	}

	switch algorithm {
	// case "tinyslam":
	// 	sc.slam = tinyslam.MakeTinySlam(robot)
	case "hectorslam":
		sc.slam = hector.MakeHectorSlam()
	default:
		return errors.New("No such SLAM algorithm.")
	}

	sc.SetState(STOPPED)

	return nil
}

// Start with a stored map
func (sc *SlamController) InitializeSlamFromStoredMap(filename, algorithm string, robot model.Robot) error {
	//Debug: fmt.Println("Entering InitializeSlamFromStoredMap")
	if sc.slam != nil {
		return errors.New("SLAM already initialized.")
	}
	//Debug: fmt.Println(filename) works okay
	mapdata, err := mapstorage.Load(filename)
	if err != nil {
		return err
	}

	switch algorithm {
	case "tinyslam":
		return errors.New("Initialize from stored map not implemented for TinySLAM.")
	case "hectorslam":
		sc.slam = hector.MakeHectorSlamFromMapRep(mapdata.MapRep)
	default:
		return errors.New("No such SLAM algorithm.")
	}

	sc.SetState(STOPPED)
	//Debug: fmt.Println("Got thru this function") seems to work to here
	return nil
}

// Terminate the slam algorithm, put the controller in the OFF state
func (sc *SlamController) TerminateSlam() {
	sc.slam = nil
	sc.SetState(OFF)

	// Run a garbage collection
	runtime.GC()
}

// Start an already initialized SLAM algorithm
func (sc *SlamController) StartSlam() error {
	if sc.slam == nil {
		return errors.New("No SLAM algorithm initialized")
	}

	sc.slam.Start()
	sc.SetState(RUNNING)

	return nil
}

// Stop a running SLAM algorithm
func (sc *SlamController) StopSlam() error {
	if sc.slam == nil {
		return errors.New("No SLAM algorithm initialized")
	}

	sc.slam.Stop()
	sc.SetState(STOPPED)

	return nil
}

// Get the SLAM algorithm
func (sc *SlamController) GetSlam() Slam {
	return sc.slam
}

// Get the map representation
func (sc *SlamController) GetMapRepresentation() maprep.MapRepresentation {
	return sc.slam.GetMapRepresentation()
}
