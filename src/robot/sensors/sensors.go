package sensors

import (
	"errors"

	"robot/config"
	"robot/fsm"
	"robot/sensors/lidar"
	"robot/sensors/logging"
	"robot/sensors/logreader"
	"robot/sensors/odometry"
	"robot/sensors/sensor"
)

const (
	SENSE   = "SENSE"
	LOGREAD = "LOGREAD"
)

// A SensorController provides a unified interface to all the sensors on the
// robot, plus logs and log reading utilities.
type SensorController struct {
	fsm.FSM
	Sensors           []sensor.Sensor
	Logger            *logging.SensorLogger
	LogReader         *logreader.SensorLogReader
	logReaderStopChan chan bool
}

// Make the default sensor controller
func MakeDefaultSensorController() *SensorController {
	sc := &SensorController{
		Sensors: make([]sensor.Sensor, 0),
		FSM:     *fsm.MakeFSM(SENSE),
	}

	if config.USE_LIDAR {
		sc.Sensors = append(sc.Sensors, lidar.LidarSensor)
	}

	if config.USE_ODOMETRY {
		sc.Sensors = append(sc.Sensors, odometry.OdometrySensor)
	}

	return sc
}

func (sc *SensorController) getSensorFromTypeName(typeName string) (sensor.Sensor, error) {
	for i := range sc.Sensors {
		if sc.Sensors[i].GetTypeName() == typeName {
			return sc.Sensors[i], nil
		}
	}

	return nil, errors.New("No such sensor.")
}

// Check if any sensor is connected
func (sc *SensorController) AnySensorConnected() bool {
	for i := range sc.Sensors {
		if sc.Sensors[i].GetState() != sensor.OFF {
			return true
		}
	}
	return false
}

// Connect a specific sensor, identified by its type name
func (sc *SensorController) ConnectSensor(typeName string) error {
	if sc.GetState() == LOGREAD {
		return errors.New("Can't connect sensor while in logreading mode")
	}

	sensor, err := sc.getSensorFromTypeName(typeName)
	if err != nil {
		return err
	}

	err = sensor.Connect()
	if sc.Logger == nil {
		sc.Logger, _ = logging.MakeTimestampedLog() // Disregard errors
	}
	sc.Logger.LogSensor(sensor)

	return err
}

// Disconnect a specific sensor, identified by its type name
func (sc *SensorController) DisconnectSensor(typeName string) error {
	sensor, err := sc.getSensorFromTypeName(typeName)
	if err != nil {
		return err
	}

	sensor.Disconnect()

	if !sc.AnySensorConnected() {
		// Terminate log file
		sc.Logger.Close()
		sc.Logger = nil
	}

	return nil
}

func (sc *SensorController) StartLogReadRealtime(logName string) error {
	if sc.AnySensorConnected() {
		return errors.New("Cannot start a log when a sensor is connected.")
	}

	var err error
	sc.LogReader, err = logreader.MakeLogReaderFromLogName(logName)
	if err != nil {
		return err
	}

	sc.SetState(LOGREAD)
	sc.logReaderStopChan, err = sc.LogReader.RealTimeReadingDistribution(true)

	return err
}

func (sc *SensorController) StopLogReadRealtime() {
	// Stop the routine, but if it's closed, continue
	select {
	case sc.logReaderStopChan <- true:
	default:
	}

	sc.LogReader.Close()

	sc.SetState(SENSE)
}

// Start a specific sensor, identified by its type name
func (sc *SensorController) StartSensor(typeName string) error {
	sensor, err := sc.getSensorFromTypeName(typeName)
	if err != nil {
		return err
	}

	return sensor.Start()
}

// Stop a specific sensor, identified by its type name
func (sc *SensorController) StopSensor(typeName string) error {
	sensor, err := sc.getSensorFromTypeName(typeName)
	if err != nil {
		return err
	}

	sensor.Stop()
	return nil
}

//func (sc *SensorController) ConnectAll() {
//
//	if sc.GetState() == LOGREAD {
//		return
//	}
//
//	for i := range sc.Sensors {
//		if sc.Sensors[i].GetState() == sensor.OFF {
//			sc.Sensors[i].Connect()
//		}
//	}
//}

func (sc *SensorController) DisconnectAll() {
	for i := range sc.Sensors {
		sc.Sensors[i].Disconnect()
	}
}
