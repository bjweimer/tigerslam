package maprep

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/skelterjohn/go.matrix"

	"hectormapping/datacontainer"
	"hectormapping/map/cache"
	"hectormapping/map/gridmap"
	"hectormapping/map/gridmap/logoddsmap"
	"hectormapping/map/gridmap/occbase"
	"hectormapping/map/mapproccontainer"
	"hectormapping/scanmatcher"
)

type MapRepSingleMap struct {
	mapContainer *mapproccontainer.MapProcContainer
}

func MakeMapRepSingleMap(mapResolution float64, mapSizeX, mapSizeY int, startCoords [2]float64) *MapRepSingleMap {
	mrsm := new(MapRepSingleMap)

	resolution := [2]int{mapSizeX, mapSizeY}

	totalMapSizeX := mapResolution * float64(mapSizeX)
	//	logger.Printf("totalMapSizeX = %f", totalMapSizeX)
	mid_offset_x := totalMapSizeX * startCoords[0]

	totalMapSizeY := mapResolution * float64(mapSizeY)
	mid_offset_y := totalMapSizeY * startCoords[1]

	gridMap := logoddsmap.MakeOccGridMapLogOdds(mapResolution, resolution, [2]float64{mid_offset_x, mid_offset_y})
	cacheMethod := cache.MakeGridMapCacheArray()
	gridMapUtil := occbase.MakeOccGridMapUtil(gridMap, cacheMethod)
	scanMatcher := scanmatcher.MakeScanMatcher()

	mrsm.mapContainer = mapproccontainer.MakeMapProcContainer(gridMap, gridMapUtil, scanMatcher)

	return mrsm
}

func (mrsm *MapRepSingleMap) Reset() {
	mrsm.mapContainer.Reset()
}

func (mrsm *MapRepSingleMap) GetScaleToMap() float64 {
	return mrsm.mapContainer.GetScaleToMap()
}

func (mrsm *MapRepSingleMap) GetMapLevels() int {
	return 1
}

func (mrsm *MapRepSingleMap) GetGridMap(mapLevel int) gridmap.OccGridMap {
	return mrsm.mapContainer.GetGridMap()
}

func (mrsm *MapRepSingleMap) AddMapMutex(i int, mapMutex *sync.RWMutex) {
	mrsm.mapContainer.AddMapMutex(mapMutex)
}

func (mrsm *MapRepSingleMap) GetMapMutex(i int) *sync.RWMutex {
	return mrsm.mapContainer.GetMapMutex()
}

func (mrsm *MapRepSingleMap) OnMapUpdated() {
	mrsm.mapContainer.ResetCachedData()
}

func (mrsm *MapRepSingleMap) MatchData(beginEstimateWorld [3]float64, dataContainer *datacontainer.DataContainer, covMatrix *matrix.DenseMatrix) [3]float64 {
	t := mrsm.mapContainer.MatchData(beginEstimateWorld, dataContainer, covMatrix, 20)
	return t
}

func (mrsm *MapRepSingleMap) UpdateByScan(dataContainer *datacontainer.DataContainer, robotPoseWorld [3]float64) {
	mrsm.mapContainer.UpdateByScan(dataContainer, robotPoseWorld)
}

func (mrsm *MapRepSingleMap) SetUpdateFactorFree(freeFactor float64) {
	mrsm.mapContainer.GetGridMap().SetUpdateFreeFactor(freeFactor)
}

func (mrsm *MapRepSingleMap) SetUpdateFactorOccupied(occupiedFactor float64) {
	mrsm.mapContainer.GetGridMap().SetUpdateOccupiedFactor(occupiedFactor)
}

// Gob Encode
func (mrsm *MapRepSingleMap) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	// Encode the gridmap
	err := encoder.Encode(mrsm.GetGridMap(0))
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Gob Decode
func (mrsm *MapRepSingleMap) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	// Decode the map
	loaded := new(logoddsmap.OccGridMapLogOdds)
	err := decoder.Decode(loaded)
	if err != nil {
		return err
	}

	cacheMethod := cache.MakeGridMapCacheArray()
	gridMapUtil := occbase.MakeOccGridMapUtil(loaded, cacheMethod)
	scanMatcher := scanmatcher.MakeScanMatcher()

	mrsm.mapContainer = mapproccontainer.MakeMapProcContainer(loaded, gridMapUtil, scanMatcher)

	return nil
}
