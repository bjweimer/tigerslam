package binarymap

import (
    "testing"
    
    hm "robot/slam/hector/hectormapping"
)

func TestShrunkenOccMap(t *testing.T) {
	
	// Make a OccGridMapLogOdds
	occMap := hm.MakeOccGridMapLogOdds(0.025, [2]int{1024, 1024}, [2]float64{25.6, 25.6})
	
	// Make a 8x shrunken Binary
	binMap := BinaryMapFromOccGridMap(occMap, 8)
	
	// Print new map parameters
	t.Logf("SizeX: %d -> %d\n", occMap.GetSizeX(), binMap.GetSizeX())
	t.Logf("SizeY: %d -> %d\n", occMap.GetSizeY(), binMap.GetSizeY())
	t.Logf("ScaleToMap: %f -> %f\n", occMap.GetScaleToMap(), binMap.GetScaleToMap())
	t.Logf("Total map sizeX: %f -> %f\n", float64(occMap.GetSizeX())*occMap.GetCellLength(), float64(binMap.GetSizeX())*binMap.GetCellLength())
	t.Logf("Total map sizeY: %f -> %f\n", float64(occMap.GetSizeY())*occMap.GetCellLength(), float64(binMap.GetSizeY())*binMap.GetCellLength())

	if (float64(occMap.GetSizeX()) * occMap.GetCellLength() != float64(binMap.GetSizeX()) * binMap.GetCellLength()) ||
			(float64(occMap.GetSizeY()) * occMap.GetCellLength() != float64(binMap.GetSizeY()) * binMap.GetCellLength()) {
		
		t.Errorf("Map lengths were changed.")
		
	}
}

