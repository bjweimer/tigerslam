// The Logreader is responsible for reading a log and distributing its data to
// the rest of the program.
package logreader

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"time"

	"robot/config"
	"robot/logging"
	"robot/sensors/lidar"
	"robot/sensors/odometry"
	"robot/sensors/sensor"
)

var logger *log.Logger
var lidarReadingType = reflect.TypeOf(&lidar.LidarReading{})
var odometryReadingType = reflect.TypeOf(&odometry.OdometryReading{})

type SensorLogReader struct {
	*csv.Reader
	file     *os.File
	fileName string
}

type LogRecord []string

func init() {
	logger = logging.New()
}

// Open log with filename
func MakeLogReaderFromLogName(logName string) (*SensorLogReader, error) {
	l := new(SensorLogReader)

	// Open the file
	f, err := os.Open(config.SENSORLOGS_ROOT + logName)
	if err != nil {
		return nil, err
	}
	l.file = f

	// Initiate the reader
	l.Reader = csv.NewReader(f)
	l.Reader.FieldsPerRecord = -1 // Don't check number of records per file

	l.fileName = logName

	return l, err
}

// Close log reader
func (slr *SensorLogReader) Close() error {
	return slr.file.Close()
}

// Read a record from the log
func (slr *SensorLogReader) ReadRecord() (LogRecord, error) {
	strings, err := slr.Read()
	if err != nil {
		return nil, err
	}
	record := LogRecord(strings)
	return record, nil
}

// Get the file name
func (slr *SensorLogReader) GetFileName() string {
	return slr.fileName
}

func (slr *SensorLogReader) ReadSensorReading() (sensorReading sensor.SensorReading, err error) {
	record, err := slr.ReadRecord()
	if err != nil {
		return
	}
	return record.GetReading()
}

// Read all records
func (slr *SensorLogReader) ReadAllRecords() ([]LogRecord, error) {
	var err error
	var record LogRecord
	records := make([]LogRecord, 0)

	for {
		record, err = slr.ReadRecord()
		if err != nil {
			break
		}
		records = append(records, record)
	}

	if err == io.EOF {
		err = nil
	}

	return records, err
}

// Read all records and convert them to sensor readings
func (slr *SensorLogReader) ReadAllSensorReadings() (sensorReadings []sensor.SensorReading, err error) {
	records, err := slr.ReadAllRecords()
	if err != nil {
		return
	}

	sensorReadings = make([]sensor.SensorReading, len(records))
	for i := range records {
		sensorReadings[i], err = records[i].GetReading()
		if err != nil {
			return
		}
	}

	return
}

// Distribute all readings from the log at the time they were taken in the log,
// relative to the start time of this function. Can be used to re-create a
// logged scenario in real-time. The flag newTimestamp decides if the
// distributed readings should be given a new timestamp.
func (slr *SensorLogReader) RealTimeReadingDistribution(newTimestamp bool) (stopChan chan bool, err error) {
	//	sensorReadings, err := slr.ReadAllSensorReadings()
	//	if err != nil {
	//		return
	//	}
	//
	//	// Return if we have no readings to distribute
	//	if len(sensorReadings) == 0 {
	//		err = errors.New("No sensor readings in log")
	//		return
	//	}

	// Read the first sensor reading of the log, distribute it and note delta
	firstReading, err := slr.ReadSensorReading()
	if err != nil {
		return
	}
	slr.Distribute(firstReading)
	deltaTime := time.Since(firstReading.GetTimestamp())

	// Delay distribution of each reading with at least the time relative to
	// start time
	var reading sensor.SensorReading
	stopChan = make(chan bool)
	go func() {
		for {
			// Check if we should stop
			select {
			case <-stopChan:
				return
			default:
				// nothing
			}

			reading, err = slr.ReadSensorReading()
			if err != nil {
				logger.Println(err)
				if err == io.EOF {
					return
				}
				return
			}

			waitTime := deltaTime - time.Since(reading.GetTimestamp())
			<-time.After(waitTime)

			// Give a new timestamp
			if newTimestamp {
				//			sensorReadings[i].SetTimestamp(sensorReadings[i].GetTimestamp().Add(timeDelta))
				reading.SetTimestamp(time.Now())
			}

			// Distribute
			err = slr.Distribute(reading)
			if err != nil {
				logger.Println(err)
			}
		}
	}()

	return
}

// Distribute a reading from the correct sensor
func (slr *SensorLogReader) Distribute(sr sensor.SensorReading) (err error) {
	switch reflect.TypeOf(sr) {
	case lidarReadingType:
		lidar.LidarSensor.Distribute(sr)
	case odometryReadingType:
		odometry.OdometrySensor.Distribute(sr)
	default:
		err = errors.New(fmt.Sprintf("Invalid reading type %s", reflect.TypeOf(sr)))
	}
	return
}

// Extract header part of record
func (r LogRecord) GetHeader() (sensorType string, timestamp time.Time, err error) {

	// Check if we have enough fields
	if len(r) < 2 {
		err = errors.New("Invalid header.")
		return
	}

	// Extract sensorType
	sensorType = r[0]

	// Extract timestamp
	timestamp, err = time.Parse(config.SENSORLOGS_TIME_FORMAT, r[1])

	return
}

// Extract body (sensordata) part of record
func (r LogRecord) GetBody() (body []string, err error) {
	if len(r) < 2 {
		err = errors.New("No body")
		return
	}

	body = r[2:]

	return

	return
}

// SensorReading from record
func (r LogRecord) GetReading() (reading sensor.SensorReading, err error) {
	sensorType, timestamp, err := r.GetHeader()
	if err != nil {
		return
	}

	body, err := r.GetBody()
	if err != nil {
		return
	}

	switch sensorType {
	case lidar.LidarSensor.GetTypeName():
		reading, err = lidar.MakeReadingFromRecordBody(body)
	case odometry.OdometrySensor.GetTypeName():
		reading, err = odometry.MakeReadingFromRecordBody(body)
	default:
		err = errors.New("Invalid sensor type")
	}
	if err != nil {
		return
	}

	reading.SetTimestamp(timestamp)

	return
}
