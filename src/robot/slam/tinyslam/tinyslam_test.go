package tinyslam

import (
    "testing"
    "os"
    "image/png"
    "fmt"
    "math"
    
    "robot/model"
    "robot/sensors/lidar"
    "robot/sensors/logreader"
)

func TestMapLaserRay(t *testing.T) {
	ts := MakeTinySlam(model.MakeDefaultDifferentialWheeledRobot())
	
	for i := 0; i < 1; i++ {
		ts.gridMap.MapLaserRay(200, 200, 600, 200, 580, 200, 100, OBSTACLE)
	}
	
	// Save image
	file, err := os.Create("testoutput/TestMapLaserRay.png")
	if err != nil {
		t.Error(err)
	}
	
	img, _ := ts.GetMapImage()
	png.Encode(file, img)
}

func TestMapUpdate(t *testing.T) {
	ts := MakeTinySlam(model.MakeDefaultDifferentialWheeledRobot())
	ts.position = model.Position{5, 17, 0}
	
	log, err := logreader.MakeLogReaderFromLogName("intel")
	if err != nil {
		t.Error(err)
	}
	
	for i := 0; i < 80; i++ {
		ts.position.X += 0.01
		
		sensorReading, err := log.ReadSensorReading()
		if err != nil {
			t.Error(err)
		}
		lidarReading, ok := sensorReading.(*lidar.LidarReading)
		if !ok {
			continue
		}
		
		cart := makeCartesianLidarReading(*lidarReading)
		ts.gridMap.MapUpdate(cart, &ts.position, 50, 100)
	}
	
	// Save image
	file, err := os.Create("testoutput/TestMapUpdate.png")
	if err != nil {
		t.Error(err)
	}
	
	img, _ := ts.GetMapImage()
	png.Encode(file, img)
}

func TestDistanceScanToMap(t *testing.T) {
	ts := MakeTinySlam(model.MakeDefaultDifferentialWheeledRobot())
	ts.position = model.Position{5, 17, 0}
	
	// Obtain a reading
	log, err := logreader.MakeLogReaderFromLogName("intel")
	if err != nil {
		t.Error(err)
	}
	sensorReading, err := log.ReadSensorReading()
	if err != nil {
		t.Error(err)
	}
	lidarReading := sensorReading.(*lidar.LidarReading)
	
	cart := makeCartesianLidarReading(*lidarReading)
	
	for i := 0; i < 50; i++ {
		// Calculate distance
		distance := ts.gridMap.DistanceCartToMap(cart, ts.position)
		t.Logf("Distance iteration %d: %d", i, distance)
		
		// Write the values to the map (should lead to lower dist next iteration)
		ts.gridMap.MapUpdate(cart, &ts.position, 50, 350)
	}
}

func TestMonteCarloSearch(t *testing.T) {
	ts := MakeTinySlam(model.MakeDefaultDifferentialWheeledRobot())
	ts.position = model.Position{5, 17, 0}
	
	// Obtain a reading
	log, err := logreader.MakeLogReaderFromLogName("intel")
	if err != nil {
		t.Error(err)
	}
	sensorReading, err := log.ReadSensorReading()
	if err != nil {
		t.Error(err)
	}
	lidarReading := sensorReading.(*lidar.LidarReading)
	
	// Write the reading to the map
	cart := makeCartesianLidarReading(*lidarReading)
	ts.gridMap.MapUpdate(cart, &ts.position, 50, 350)
	
	// Find position using the same reading using monte carlo search, using a
	// slightly distorted assumed position.
	position := model.Position{
		ts.position.X + 0.15,
		ts.position.Y - 0.55,
		ts.position.Theta + 0.05,
	}
	bestpos := ts.gridMap.monteCarloSearch(cart, &position, 0.5, 0.5, 1000, nil)
	
	deltapos := model.Position{
		ts.position.X - bestpos.X,
		ts.position.Y - bestpos.Y,
		ts.position.Theta - bestpos.Theta,
	}
	
	t.Logf("Best position: %s", bestpos)
	t.Logf("Delta: %s", deltapos)	
}

func saveImage(filename string, ts *TinySlam) {
	file, err := os.Create("testoutput/" + filename + ".png")
	if err != nil {
		return
	}
	
	img, _ := ts.GetMapImage()
	png.Encode(file, img)
	file.Close()
}

func TestLidarOnlyTinySlam(t *testing.T) {
	ts := MakeTinySlam(model.MakeDefaultDifferentialWheeledRobot())
	ts.position = model.Position{25, 37, 0}
	
	// Start a logreader
	log, err := logreader.MakeLogReaderFromLogName("intel")
	if err != nil {
		t.Error(err)
	}
	
	for i := 0; i < 2000; i++ {
	
		// Obtain lidar reading
		sensorReading, err := log.ReadSensorReading()
		if err != nil {
			t.Error(err)
		}
		lidarReading, ok := sensorReading.(*lidar.LidarReading)
		if !ok {
			continue	// It wasn't a lidar reading
		}
		
		// Try to estimate new position
		estPos := &model.Position{ts.position.X, ts.position.Y, ts.position.Theta}
		if ts.velocity < 5.0 {
			deltaTime := lidarReading.GetTimestamp().Sub(ts.timestamp)
			estPos.Theta += ts.thetadot * deltaTime.Seconds()
			estPos.X += ts.velocity * math.Cos(estPos.Theta) * deltaTime.Seconds()
			estPos.Y += ts.velocity * math.Sin(estPos.Theta) * deltaTime.Seconds()
		}
		
		ts.LidarMapBuilding(lidarReading, &ts.position)
		
		if i % 50 == 0 {
			saveImage(fmt.Sprintf("TestLidarOnlyTinySlam-it%d", i), ts)
		}
	}
	
	saveImage("TestLidarOnlyTinySlam", ts)
	
}

func BenchmarkMapLaserRay(b *testing.B) {
	b.StopTimer()
	ts := MakeTinySlam(model.MakeDefaultDifferentialWheeledRobot())
	b.StartTimer()
	
	for i := 0; i < b.N; i++ {
		ts.gridMap.MapLaserRay(200, 200, 600, 200, 550, 200, 50, OBSTACLE)
	}
}