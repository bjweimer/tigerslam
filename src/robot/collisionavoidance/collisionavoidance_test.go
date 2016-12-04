package collisionavoidance

import (
	"fmt"
	"sync"
	"testing"

	"robot/sensors/lidar"
)

func TestCollissionAvoidance(t *testing.T) {

	err := lidar.LidarSensor.Connect()
	if err != nil {
		t.Fatal(err)
	}

	err = lidar.LidarSensor.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer lidar.LidarSensor.Stop()
	defer lidar.LidarSensor.Disconnect()

	cd := MakeDefaultCollisionDetector()

	n := 10
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {

		for i := 0; i < n; i++ {
			select {
			case <-cd.StopChan:
				fmt.Println("Stop!")
			case <-cd.ResumeChan:
				fmt.Println("Resume")
			}
		}

		wg.Done()

	}()

	cd.Start()
	wg.Wait()

	cd.Stop()

}
