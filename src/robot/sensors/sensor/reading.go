package sensor

import (
	"time"
	"fmt"
	
	"robot/config"
)

type SensorReading interface {
	GetSensor() Sensor
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
	String() string
	LogEntry() string
}

type BasicSensorReading struct {
	Sensor Sensor
	timestamp time.Time
}

func (bsr *BasicSensorReading) GetSensor() Sensor {
	return bsr.Sensor
}

func (bsr *BasicSensorReading) GetTimestamp() time.Time {
	return bsr.timestamp
}

func (bsr *BasicSensorReading) SetTimestamp(timestamp time.Time) {
	bsr.timestamp = timestamp
}

func (bsr *BasicSensorReading) String() string {
	return bsr.GetSensor().GetTypeName() + " reading at " + bsr.timestamp.Format(time.StampNano)
}

func (bsr *BasicSensorReading) LogEntryHeader() string {
	return fmt.Sprintf("%s,%s", bsr.GetSensor().GetTypeName(), bsr.timestamp.Format(config.SENSORLOGS_TIME_FORMAT))
}

func (bsr *BasicSensorReading) LogEntry() string {
	return bsr.LogEntryHeader() + ",NOIMPLEMENTEDSENSORDATA"
}