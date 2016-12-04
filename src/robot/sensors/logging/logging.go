package logging

import (
	"os"
	"time"
	"log"
	
	"robot/config"
	"robot/sensors/sensor"
)

type SensorLogger struct {
	*log.Logger
	filename string
	file *os.File
	stopChans []chan bool
}

// Make a log of arbitrary filename
func MakeLog(filename string) (*SensorLogger, error) {
	l := new(SensorLogger)
	
	l.filename = filename
	f, err := os.Create(config.SENSORLOGS_ROOT + filename + ".log")
	if err != nil {
		return nil, err
	}
	l.file = f
	
	logger := log.New(f, "", 0)
	l.Logger = logger
	
	l.stopChans = make([]chan bool, 0)
	
	return l, nil
}

// Make a log with timestamped filename
func MakeTimestampedLog() (*SensorLogger, error) {
	return MakeLog(time.Now().Format("01-02-2006.15-04-05"))
}

// Write a sensor reading to the log
func (l *SensorLogger) writeReading(r sensor.SensorReading) {
	l.Println(r.LogEntry())
}

// Log a sensor
func (l *SensorLogger) LogSensor(s sensor.Sensor) {
	var reading sensor.SensorReading
	var ok bool
	
	stopChan := make(chan bool)
	l.stopChans = append(l.stopChans, stopChan)
		
	ch := s.Subscribe()
	go func() {
		for {
			select {
			case reading, ok = <-ch:
				// continue
			case <-stopChan:
				s.Unsubscribe(ch)
				return
			}
			
			// End if channel is closed
			if !ok {
				return
			}
			
			l.writeReading(reading)
		}
	}()
}

// Get the file name
func (l *SensorLogger) GetFileName() string {
	return l.filename
}

// Close
func (l *SensorLogger) Close() error {
	for i := range l.stopChans {
		l.stopChans[i] <- true
	}
	l.stopChans = make([]chan bool, 0)
	return l.file.Close()
}