package lidar

import (
    "testing"
    "time"
    
    "robot/sensors/sensor"
)

func TestLidarDistribution(t *testing.T) {
	ch := LidarSensor.Subscribe()
	
	n := 5
	
	// Send n readings
	go func() {
		for i := 0; i < n; i++ {
			time.Sleep(100 * time.Millisecond)
			reading := MakeLidarReading()
			reading.SetTimestamp(time.Now())
			LidarSensor.Distribute(reading)
		}
	}()
	
	// Receive readings
	for i := 0; i < n; i++ {
		reading := <-ch
		t.Log(reading)
	}
}

func TestTypeAssertion(t *testing.T) {
	var sensorReading sensor.SensorReading
	sensorReading = MakeLidarReading()
	t.Log(sensorReading)
	
	lidarReading := (sensorReading).(*LidarReading)
	t.Log(lidarReading)
}

func TestActualLidar(t *testing.T) {
	ch := LidarSensor.Subscribe()
	
	n := 2
	
	// Receive readings
	go func() {
		for i := 0; i < n; i++ {
			reading := <-ch
			t.Log(reading.LogEntry())
		}
		LidarSensor.Stop()
	}()
		
	LidarSensor.Start()
}