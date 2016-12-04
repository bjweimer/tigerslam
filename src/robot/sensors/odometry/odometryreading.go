package odometry

import (
	"fmt"
	"strconv"

	"robot/sensors/sensor"
)

// An OdometryReading consists of the standard sensor reading parameters, plus
// the increment done on the left and right wheels (integer).
type OdometryReading struct {
	sensor.BasicSensorReading
	LeftPulses  int
	RightPulses int
}

// Construct a plain record
func MakeOdometryReading() *OdometryReading {
	or := &OdometryReading{
		BasicSensorReading: sensor.BasicSensorReading{
			Sensor: OdometrySensor,
		},
	}

	return or
}

// Given the data part of a log record, construct a reading with the data from
// the record.
func MakeReadingFromRecordBody(recordData []string) (or *OdometryReading, err error) {
	or = MakeOdometryReading()

	left, err := strconv.ParseInt(recordData[0], 10, 64)
	if err != nil {
		return or, err
	}

	right, err := strconv.ParseInt(recordData[1], 10, 64)
	if err != nil {
		return or, err
	}

	or.LeftPulses = int(left)
	or.RightPulses = int(right)

	return or, nil
}

// Construct the data part of the log entry
func (or *OdometryReading) LogEntryData() string {
	return fmt.Sprintf(",%d,%d", or.LeftPulses, or.RightPulses)
}

// Construct the full log entry
func (or *OdometryReading) LogEntry() string {
	return or.LogEntryHeader() + or.LogEntryData()
}
