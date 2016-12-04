package logreader

import (
    "testing"
    "time"
    "fmt"
    
    "robot/sensors/lidar"
)

// Test if we can make a logreader from the testlog
func TestMakeLogReaderFromLogName(t *testing.T) {
	_, err := MakeLogReaderFromLogName("testlog")
	if err != nil {
		t.Error(err)
	}
}

// Read all records of the testlog
func TestReadAllRecords(t *testing.T) {
	logReader, err := MakeLogReaderFromLogName("testlog")
	if err != nil {
		t.Fatal(err)
	}
	
	allRecords, err := logReader.ReadAllRecords()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(allRecords) == 0 {
		t.Fatal("Length of all records was 0")
	}
	
	if len(allRecords[0]) < 2 {
		t.Errorf("Length of first record was %d (not enough for header)", len(allRecords[0]))
	}
	
	t.Logf("Read %d records.\n", len(allRecords))
	t.Logf("First record: %s", allRecords[0])
}

// Read all records of the testlog as recreated sensor readings
func TestReadAllSensorReadings(t *testing.T) {
	logReader, err := MakeLogReaderFromLogName("testlog")
	if err != nil {
		t.Fatal(err)
	}
	
	allReadings, err := logReader.ReadAllSensorReadings()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(allReadings) == 0 {
		t.Fatal("Length of all readings was 0")
	}
	
	for i := range allReadings {
		t.Logf("Reading %d: %s", i, allReadings[i])
	}
	
	sinceFirst := time.Since(allReadings[0].GetTimestamp())
	t.Logf("Time since log began: %.1f hours", sinceFirst.Hours())
}

// Read all records of the testlog and distribute them with time delay according
// to the log
func TestRealTimeReadingDistribution(t *testing.T) {
	logReader, err := MakeLogReaderFromLogName("testlog")
	if err != nil {
		t.Fatal(err)
	}
	
	// Distribute the readings from the log
	go func() {
		err = logReader.RealTimeReadingDistribution(true)
		if err != nil {
			t.Fatal(err)
		}
		lidar.LidarSensor.CloseSubscribtions()
	}()
	
	// Receive the sensor distributions from LIDAR
	var firstTimestamp time.Time
	ch := lidar.LidarSensor.Subscribe()
		
	count := 0
	for {
		reading, ok := <-ch
		if !ok {
			break
		}
		
		if count == 0 {
			firstTimestamp = reading.GetTimestamp()
		}
		count++
		
		t.Logf("Received reading: %s", reading)
	}
	
	distributedDuration := time.Since(firstTimestamp)
		
	t.Logf("Distribution duration %f seconds", distributedDuration.Seconds())
}

// Read a "real" (i.e. very long) data sets' records and distribute them with
// time delay
func TestRealTimeReadingDistributionLongDataSet(t *testing.T) {
	logReader, err := MakeLogReaderFromLogName("intel")
	if err != nil {
		t.Fatal(err)
	}
	
	// Distribute the readings from the log
	go func() {
		err = logReader.RealTimeReadingDistribution(true)
		if err != nil {
			t.Error(err)
		}
		lidar.LidarSensor.CloseSubscribtions()
	}()
	
	// Receive the sensor distributions from LIDAR
	ch := lidar.LidarSensor.Subscribe()
	count := 0
	for {
		reading, ok := <-ch
		if !ok {
			break
		}
		count++
		
		// Convert to LidarReading
		lidarReading := reading.(*lidar.LidarReading)
		
		// Use fmt instead of log, so we get output as we go
		fmt.Printf("%d: %s\n", count, lidarReading)
		fmt.Printf("Send-receive delay: %.3f s\n", time.Since(lidarReading.GetTimestamp()).Seconds())
	}
}