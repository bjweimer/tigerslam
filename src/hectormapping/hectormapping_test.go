package hectormapping

import (
    "testing"
    "os"
    "image"
    "image/color"
    "image/png"
    "fmt"
    
    "robot/sensors/logreader"
    "robot/sensors/lidar"
    "robot/config"
    "robot/logging"
)

func saveImage(im image.Image, name string, t *testing.T) {
	// Save image
	file, err := os.Create("testoutput/" + name + ".png")
	if err != nil {
		t.Error(err)
	}
	err = png.Encode(file, im)
	if err != nil {
		t.Error(err)
	}
}

func saveImageFromGridMap(gridMap OccGridMap, name string, t *testing.T) {
	im := image.NewGray(image.Rect(0, 0, gridMap.GetSizeX(), gridMap.GetSizeY()))
	
	// Declare colors
	unknownColor := color.Gray{127}
	freeColor := color.Gray{255}
	occupiedColor := color.Gray{0}
	
	// Fill image
	for x := 0; x < gridMap.GetSizeX(); x++ {
		for y := 0; y < gridMap.GetSizeY(); y++ {
			
			cell := gridMap.GetCell(x, y)
			
			if cell.IsFree() {
				im.Set(x, y, freeColor)
			} else if cell.IsOccupied() {
				im.Set(x, y, occupiedColor)
			} else {
				im.Set(x, y, unknownColor)
			}
			
		}
	}
	
	saveImage(im, name, t)
}

// Make images such as Fig. 2 in the article, showing the partial derivatives
// of the map, i.e. ∂M/∂x and ∂M/∂Y. 
func getDerivativeImages(hsp *HectorSlamProcessor, t *testing.T) {
	// Get MapProcContainer and gridmaputil
	mapProcContainer := hsp.GetMapRepresentation().GetMapProcContainer()
	gridMapUtil := mapProcContainer.GetGridMapUtil()
	gridMap := mapProcContainer.GetGridMap()
	
	logger.Printf("Obstacle threshold: %f", gridMap.GetObstacleThreshold())
	
	zoomPx := 10
	
	var xMax, yMax, xMin, yMin int
	gridMap.GetMapExtends(&xMax, &yMax, &xMin, &yMin)
	t.Logf("xMax = %d, yMax = %d, xMin = %d, yMin = %d", xMax, yMax, xMin, yMin)
	
	imInter := image.NewGray(image.Rect(xMin * zoomPx, yMin * zoomPx, xMax * zoomPx, yMax * zoomPx))
	imdMdx := image.NewGray(image.Rect(xMin * zoomPx, yMin * zoomPx, xMax * zoomPx, yMax * zoomPx))
	imdMdy := image.NewGray(image.Rect(xMin * zoomPx, yMin * zoomPx, xMax * zoomPx, yMax * zoomPx))
	
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			
			for xi := 0; xi < zoomPx; xi++ {
				for yi := 0; yi < zoomPx; yi++ {
					floatMapCoords := [2]float64{float64(x) + float64(xi)/float64(zoomPx), float64(y) + float64(yi)/float64(zoomPx)}
					r := gridMapUtil.InterpMapValueWithDerivatives(floatMapCoords)
			
					imInter.Set(x*zoomPx + xi, y*zoomPx + yi, color.Gray{uint8(r[0] * 255)})
					imdMdx.Set(x*zoomPx + xi, y*zoomPx + yi, color.Gray{uint8(128 + r[1] * 128)})
					imdMdy.Set(x*zoomPx + xi, y*zoomPx + yi, color.Gray{uint8(128 + r[2] * 128)})
					
				}
			}
			
		}
	}
	
	
	saveImage(imInter, "interpolated", t)
	saveImage(imdMdx, "dMdx", t)
	saveImage(imdMdy, "dMdy", t)
}

func TestHectorSlamProcessor(t *testing.T) {
	logging.AddWriter(os.Stdout)
	
//	hsp := MakeHectorSlamProcessor(0.0125, 4096, 4096, [2]float64{0.1, 0.1}, 6)	// Andre etasje gamle elektro.log
//	hsp := MakeHectorSlamProcessor(0.025, 2048, 2048, [2]float64{0.1, 0.4}, 3) // EL5
//	hsp := MakeHectorSlamProcessor(0.025, 4096, 4096, [2]float64{0.1, 0.4}, 3) // Glassgarden
	hsp := MakeHectorSlamProcessor(0.025, 4096, 4096, [2]float64{0.1, 0.4}, 3) // Infohjornerunde
//	hsp := MakeHectorSlamProcessor(0.025, 4096, 4096, [2]float64{0.1, 0.4}, 3) // Runde gamle blokker
	hsp.SetUpdateFactorFree(0.4)
	hsp.SetUpdateFactorOccupied(0.9)
	
	sensorLog, err := logreader.MakeLogReaderFromLogName("Infohjornerunde.log")
	if err != nil {
		t.Fatal(err)
	}
	
	dataContainer := MakeDataContainer(config.LIDAR_NUM_DISTANCES)
	
	n := 10000
	for i := 0; i < n; i++ {
		sensorReading, _ := sensorLog.ReadSensorReading()
		
		if sensorReading == nil {
			break
		}
		
		lidarReading, ok := sensorReading.(*lidar.LidarReading)
		if !ok {
			t.Errorf("Could not convert sensorReading to lidarReading")
		}
		
		LidarReadingToDataContainer(lidarReading, dataContainer, hsp.GetScaleToMap())
		hsp.Update(dataContainer, hsp.GetLastScanMatchPose())
		
		estimate := hsp.GetLastScanMatchPose()
		logger.Printf("%d: Estimate: %f, %f, %f", i, estimate[0], estimate[1], estimate[2])
	}
	
	// Save image
	saveImageFromGridMap(hsp.GetGridMap(), "TestHectorSlamProcessor", t)
	
	for i := 1; i < hsp.GetMapLevels(); i++ {
		saveImageFromGridMap(hsp.GetGridMapByLevel(i), fmt.Sprintf("TestHectorSlamProcessor-%d", i), t)
	}
	
//	getDerivativeImages(hsp, t)
}

