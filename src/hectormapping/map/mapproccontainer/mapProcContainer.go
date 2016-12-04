package mapproccontainer

import (
	"sync"
	
	"github.com/skelterjohn/go.matrix"
	
	"hectormapping/map/gridmap"
	"hectormapping/map/gridmap/occbase"
	"hectormapping/scanmatcher"
	"hectormapping/datacontainer"
)

type MapProcContainer struct {
	gridMap gridmap.OccGridMap
	gridMapUtil *occbase.OccGridMapUtil
	scanMatcher *scanmatcher.ScanMatcher
	mapMutex *sync.RWMutex
}

func MakeMapProcContainer(gridMap gridmap.OccGridMap, gridMapUtil *occbase.OccGridMapUtil, 
		scanMatcher *scanmatcher.ScanMatcher) *MapProcContainer {
	
	return &MapProcContainer{
		gridMap: gridMap,
		gridMapUtil: gridMapUtil,
		scanMatcher: scanMatcher,
	}
	
}

func (mpc *MapProcContainer) Cleanup() {
	mpc.gridMap = nil
	mpc.gridMapUtil = nil
	mpc.scanMatcher = nil
	mpc.mapMutex = nil
}

func (mpc *MapProcContainer) Reset() {
	mpc.gridMap.Reset()
	mpc.gridMapUtil.ResetCachedData()
}

func (mpc *MapProcContainer) ResetCachedData() {
	mpc.gridMapUtil.ResetCachedData()
}

func (mpc *MapProcContainer) GetScaleToMap() float64 {
	return mpc.gridMap.GetScaleToMap()
}

func (mpc *MapProcContainer) GetGridMap() gridmap.OccGridMap {
	return mpc.gridMap
}

func (mpc *MapProcContainer) AddMapMutex(mapMutex *sync.RWMutex) {
	mpc.mapMutex = mapMutex
}

func (mpc *MapProcContainer) GetMapMutex() *sync.RWMutex {
	return mpc.mapMutex
}

func (mpc *MapProcContainer) MatchData(beginEstimateWorld [3]float64,
		dataContainer *datacontainer.DataContainer, covMatrix *matrix.DenseMatrix,
		maxIterations int) [3]float64 {
	
	return mpc.scanMatcher.MatchData(beginEstimateWorld, mpc.gridMapUtil,
			dataContainer, covMatrix, maxIterations)
	
}

func (mpc *MapProcContainer) UpdateByScan(dataContainer *datacontainer.DataContainer, 
		robotPoseWorld [3]float64) {
	
	if mpc.mapMutex != nil {
		mpc.mapMutex.Lock()
	}
	
	mpc.gridMap.UpdateByScan(dataContainer, robotPoseWorld)
	
	if mpc.mapMutex != nil {
		mpc.mapMutex.Unlock()
	}
	
}