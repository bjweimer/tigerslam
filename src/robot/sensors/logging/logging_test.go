package logging

import (
    "testing"
    "time"
    
    "robot/sensors/lidar"
)

func TestLogData(t *testing.T) {
	logger, err := MakeTimestampedLog()
	if err != nil {
		t.Error(err)
	}
	
	n := 5
	
	// Distribute n readings
	go func() {
		for i := 0; i < n; i++ {
			time.Sleep(100 * time.Millisecond)
			reading := lidar.MakeLidarReading()
			reading.SetTimestamp(time.Now())
			lidar.LidarSensor.Distribute(reading)
		}
		lidar.LidarSensor.CloseSubscribtions()
	}()
	
	logger.LogSensor(lidar.LidarSensor)
	
}

