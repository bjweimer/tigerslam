package maprep

import (
	"bytes"
	"encoding/gob"
	"math"
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

// Map representation containing several maps, as described in the Hector SLAM
// article "A Flexible and Scalable SLAM System with Full 3D Motion
// Estimation". Each map has a different resolution, more specificly, they are
// stacked with each map having half of the resolution of the previous. They
// are simultaneously updated, and not generated from each other.
//
// There are n maps and n-1 data containers. Map 0 uses the incoming
// dataContainer directly, while the rest uses new dataContainers generated
// from the incoming one, scaled to their respective resolutions.
type MapRepMultiMap struct {
	mapContainer   []*mapproccontainer.MapProcContainer
	dataContainers []*datacontainer.DataContainer
}

// Make a MapRepMultiMap object with a given resolution (cell length in
// meters), size in x and y directions (in the number of cells), start
// coordinates on the map ranging from 0.0 to 1.0 and numDepth, the number
// of maps to use. Using numDepth > 1, a corresponding number of smaller maps
// will be initialized, updated and used for matching.
func MakeMapRepMultiMap(mapResolution float64, mapSizeX, mapSizeY int,
	numDepth int, startCoords [2]float64) *MapRepMultiMap {

	mrmm := new(MapRepMultiMap)
	mrmm.dataContainers = make([]*datacontainer.DataContainer, numDepth-1)
	mrmm.mapContainer = make([]*mapproccontainer.MapProcContainer, numDepth)

	resolution := [2]int{mapSizeX, mapSizeY}

	totalMapSizeX := mapResolution * float64(mapSizeX)
	mid_offset_x := totalMapSizeX * startCoords[0]

	totalMapSizeY := mapResolution * float64(mapSizeY)
	mid_offset_y := totalMapSizeY * startCoords[1]

	for i := 0; i < numDepth; i++ {
		//		logger.Printf("HectorSM map lvl %d: cellLength: %f res x: %f res y: %f\n",
		//			i, mapResolution, mid_offset_x, mid_offset_y)

		gridMap := logoddsmap.MakeOccGridMapLogOdds(mapResolution, resolution, [2]float64{mid_offset_x, mid_offset_y})
		cacheMethod := cache.MakeGridMapCacheArray()
		gridMapUtil := occbase.MakeOccGridMapUtil(gridMap, cacheMethod)
		scanMatcher := scanmatcher.MakeScanMatcher()

		mrmm.mapContainer[i] = mapproccontainer.MakeMapProcContainer(gridMap, gridMapUtil, scanMatcher)

		resolution[0] /= 2
		resolution[1] /= 2
		mapResolution *= 2.0
	}

	for i := range mrmm.dataContainers {
		mrmm.dataContainers[i] = datacontainer.MakeDataContainer(0)
	}

	return mrmm
}

func (mrmm *MapRepMultiMap) Reset() {
	for i := range mrmm.mapContainer {
		mrmm.mapContainer[i].Reset()
	}
}

func (mrmm *MapRepMultiMap) GetScaleToMap() float64 {
	return mrmm.mapContainer[0].GetScaleToMap()
}

func (mrmm *MapRepMultiMap) GetMapLevels() int {
	return len(mrmm.mapContainer)
}

func (mrmm *MapRepMultiMap) GetGridMap(mapLevel int) gridmap.OccGridMap {
	return mrmm.mapContainer[mapLevel].GetGridMap()
}

func (mrmm *MapRepMultiMap) AddMapMutex(i int, mapMutex *sync.RWMutex) {
	mrmm.mapContainer[i].AddMapMutex(mapMutex)
}

func (mrmm *MapRepMultiMap) GetMapMutex(i int) *sync.RWMutex {
	return mrmm.mapContainer[i].GetMapMutex()
}

func (mrmm *MapRepMultiMap) OnMapUpdated() {
	for i := range mrmm.mapContainer {
		mrmm.mapContainer[i].ResetCachedData()
	}
}

// Match the incoming dataContainer (LIDAR scan) with the maps. The matching
// starts at the coarsest level, with beginEstimateWorld as an initial guess.
// The matching then works its way up to the finest level, where the result of
// the previous (coarser) map is used as an initial guess for the current
// (finer) map.
func (mrmm *MapRepMultiMap) MatchData(beginEstimateWorld [3]float64,
	dataContainer *datacontainer.DataContainer, covMatrix *matrix.DenseMatrix) [3]float64 {

	estimate := beginEstimateWorld

	for i := len(mrmm.mapContainer) - 1; i >= 0; i-- {

		if i == 0 {
			// We're at the finest map, use the incoming dataContainer directly
			estimate = mrmm.mapContainer[i].MatchData(estimate, dataContainer, covMatrix, 5)
		} else {
			// We're at a coarser map, first make a scaled version of the
			// incoming dataContainer, then use it to match with this map.
			mrmm.dataContainers[i-1].SetFrom(dataContainer, 1.0/math.Pow(2.0, float64(i)))
			estimate = mrmm.mapContainer[i].MatchData(estimate, mrmm.dataContainers[i-1], covMatrix, 3)
		}

	}

	return estimate

	//	size := len(mrmm.mapContainer)
	//	tmp := beginEstimateWorld
	//	index := size - 1
	//
	//	for i := 0; i < size; i++ {
	//		if index == 0 {
	//			tmp = mrmm.mapContainer[index].MatchData(tmp, dataContainer, covMatrix, 5)
	//		} else {
	//			mrmm.dataContainers[index - 1].SetFrom(dataContainer, 1.0 / math.Pow(2.0, float64(index)))
	//			tmp = mrmm.mapContainer[index].MatchData(tmp, mrmm.dataContainers[index - 1], covMatrix, 3)
	//		}
	//	}
	//
	//	return tmp
}

// Update each map. This function assumes that MatchData has already been
// executed for this dataContainer, and so the maps with index > 0 (the smaller
// ones) uses their cached dataContainers.
func (mrmm *MapRepMultiMap) UpdateByScan(dataContainer *datacontainer.DataContainer, robotPoseWorld [3]float64) {
	for i := range mrmm.mapContainer {
		if i == 0 {
			mrmm.mapContainer[i].UpdateByScan(dataContainer, robotPoseWorld)
		} else {
			mrmm.mapContainer[i].UpdateByScan(mrmm.dataContainers[i-1], robotPoseWorld)
		}
	}
}

func (mrmm *MapRepMultiMap) SetUpdateFactorFree(freeFactor float64) {
	for i := range mrmm.mapContainer {
		mrmm.mapContainer[i].GetGridMap().SetUpdateFreeFactor(freeFactor)
	}
}

func (mrmm *MapRepMultiMap) SetUpdateFactorOccupied(occupiedFactor float64) {
	for i := range mrmm.mapContainer {
		mrmm.mapContainer[i].GetGridMap().SetUpdateOccupiedFactor(occupiedFactor)
	}
}

func (mrmm *MapRepMultiMap) GetMapProcContainer() *mapproccontainer.MapProcContainer {
	return mrmm.mapContainer[0]
}

// Gob Encode
func (mrmm *MapRepMultiMap) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	// Encode the number of levels we are using
	err := encoder.Encode(mrmm.GetMapLevels())
	if err != nil {
		return nil, err
	}

	// Encode all gridmaps, they are the only things we really need
	for _, mc := range mrmm.mapContainer {
		err := encoder.Encode(mc.GetGridMap())
		if err != nil {
			return nil, err
		}
	}

	return w.Bytes(), nil
}

// Gob Decode
func (mrmm *MapRepMultiMap) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	// Decode the number of levels
	var numDepth int
	err := decoder.Decode(&numDepth)
	if err != nil {
		return err
	}

	// Make the slices for dataContainers and mapContainers
	mrmm.dataContainers = make([]*datacontainer.DataContainer, numDepth-1)
	mrmm.mapContainer = make([]*mapproccontainer.MapProcContainer, numDepth)

	for i := range mrmm.dataContainers {
		mrmm.dataContainers[i] = datacontainer.MakeDataContainer(0)
	}

	// Decode each gridmap
	for i := 0; i < numDepth; i++ {

		// Decode the map itself
		loadedMap := new(logoddsmap.OccGridMapLogOdds)
		err = decoder.Decode(loadedMap)
		if err != nil {
			return err
		}

		// Set up this MapProcContainer
		cacheMethod := cache.MakeGridMapCacheArray()
		gridMapUtil := occbase.MakeOccGridMapUtil(loadedMap, cacheMethod)
		scanMatcher := scanmatcher.MakeScanMatcher()

		mrmm.mapContainer[i] = mapproccontainer.MakeMapProcContainer(loadedMap, gridMapUtil, scanMatcher)
	}

	return nil
}
