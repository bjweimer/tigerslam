package lidar

import (
	"fmt"
	"strconv"
	
	"robot/sensors/sensor"
)

//
type LidarReading struct {
	sensor.BasicSensorReading
	Distances []float64
	Span float64
	MaxDistance float64
}

// Construct a plain record
func MakeLidarReading() *LidarReading {
	l := &LidarReading {
		BasicSensorReading: sensor.BasicSensorReading {
			Sensor: LidarSensor,
		},
		Distances: make([]float64, LidarSensor.Distances),
		Span: LidarSensor.RadialSpan,
		MaxDistance: LidarSensor.MaxDistance,
	}
	
	return l
}

// Given the data part of a log record, construct a reading with the data from
// the record
func MakeReadingFromRecordBody(recordData []string) (l *LidarReading, err error) {
	l = MakeLidarReading()
	
	// The first value is the span of the lidar sweep
	a, err := strconv.ParseFloat(recordData[0], 64)
	if err != nil {
		return nil, err
	}
	l.Span = a
	
	// The rest are distances
	n := len(recordData) - 1
	if len(l.Distances) < n {
		l.Distances = make([]float64, n)
	}
	
	for i := 0; i < n; i++ {
		a, err := strconv.ParseFloat(recordData[i+1], 64)
		if err != nil {
			return nil, err
		}
		l.Distances[i] = a
	}
	
	return l, nil
}

// Construct the data part of the log entry
func (lr *LidarReading) LogEntryData() string {
	// Span
	s := fmt.Sprintf(",%f", lr.Span)
	
	// Distances
	for i := 0; i < len(lr.Distances); i++ {
		s += fmt.Sprintf(",%.0f", lr.Distances[i])
	}
	
	return s
}

// Construct the full log entry
func (lr *LidarReading) LogEntry() string {
	return lr.LogEntryHeader() + lr.LogEntryData()
}