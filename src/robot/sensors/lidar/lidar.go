// The Lidar package is responsible for communicating with the LIDAR sensor
// which the robot is equipped with.
package lidar

import (
	"errors"
	"log"
	"time"

	"robot/config"
	"robot/fsm"
	"robot/logging"
	"robot/sensors/lidar/driver"
	"robot/sensors/sensor"
)

var logger *log.Logger
var LidarSensor *Lidar

const NAME = "LIDAR" // Used to identify the sensor for i.e. web gui/logs

func init() {
	LidarSensor = MakeDefaultLidar()
	logger = logging.New()
}

// A LIDAR implements a Sensor and holds additional LIDAR specific information,
// such as the ranges and the number of distance readings it produces for every
// reading.
type Lidar struct {
	sensor.BasicSensor
	urg         *driver.Urg
	stopChan    chan bool
	RadialSpan  float64
	MaxDistance float64
	Distances   int
}

// Make an arbitrary LIDAR
func MakeLidar(radialSpan, maxDistance float64, distances int) *Lidar {
	l := &Lidar{
		BasicSensor: sensor.BasicSensor{
			FSM:  *fsm.MakeFSM(sensor.OFF),
			Name: NAME,
		},
		RadialSpan:  radialSpan,
		MaxDistance: maxDistance,
		Distances:   distances,
		urg:         driver.MakeDefaultUrg(),
	}

	l.stopChan = make(chan bool)

	return l
}

// Make LIDAR from default parameters and config file
func MakeDefaultLidar() *Lidar {
	return MakeLidar(config.LIDAR_RADIAL_SPAN, config.LIDAR_MAX_DISTANCE, config.LIDAR_NUM_DISTANCES)
}

// Return all parameters as a string-interface map
func (l Lidar) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"Radial span":  l.RadialSpan,
		"Max distance": l.MaxDistance,
		"Distances":    l.Distances,
	}
}

// Connect the LIDAR
func (l *Lidar) Connect() error {
	err := l.urg.Connect()
	if err != nil {
		return err
	}

	// Update the state
	l.SetState(sensor.CONNECTED)

	logger.Println("LIDAR connected.")

	return nil
}

// Disconnect the LIDAR
func (l *Lidar) Disconnect() {
	l.urg.Disconnect()

	// Update the state
	l.SetState(sensor.OFF)
	logger.Println("LIDAR disconnected.")
}

// Start measuring and distribution
func (l *Lidar) Start() error {

	var err error

	// Check if we're connected
	if l.GetState() != sensor.CONNECTED {
		return errors.New("LIDAR not connected.")
	}

	// Set the lidar to continuously take measurements
	err = l.urg.RequestInfiniteData()
	if err != nil {
		logger.Println(err)
		return err
	}

	l.SetState(sensor.RUNNING)
	logger.Println("LIDAR now running.")

	// Now continuously read values and distribute them
	go func() {
		for {
			select {
			case <-l.stopChan:
				l.SetState(sensor.CONNECTED) // maybe should be sensor.OFF?????
				logger.Println("LIDAR stopped.")
				return
			default:
			}

			distances, err := l.urg.ReceiveData()
			if err != nil {
				// log the error but try to continue
				logger.Println(err)

				continue
			}
			//fmt.Printf("\033[17;0H") //move cursor to row 6, col 0
			//fmt.Printf("                                                     ")
			//fmt.Printf("\033[17;0H") //display in one place
			//fmt.Printf("degree/distance %d = %f\n", 90, distances[90])

			reading := &LidarReading{
				BasicSensorReading: sensor.BasicSensorReading{
					Sensor: l,
				},
				Distances:   distances,
				Span:        config.LIDAR_RADIAL_SPAN,
				MaxDistance: l.MaxDistance,
			}
			reading.SetTimestamp(time.Now())

			l.Distribute(reading)
		}
	}()

	return nil
}

// Stop the running LIDAR. Note that this function does not alter the state
// iteself. This function communicates the stopping wish to the running go
// routine, which alters the state when it terminates.
func (l *Lidar) Stop() {
	logger.Println("Lidar stopping ...")
	l.stopChan <- true
}
