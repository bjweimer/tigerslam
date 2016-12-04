package hectormapping

import (
	"math"
	"sync"

	"github.com/skelterjohn/go.matrix"

	"hectormapping/datacontainer"
	"hectormapping/map/gridmap"
	"hectormapping/map/maprep"
	"hectormapping/utils"
)

type HectorSlamProcessor struct {
	mapRep maprep.MapRepresentation

	lastMapUpdatePose [3]float64
	lastScanMatchPose [3]float64
	lastScanMatchCov  *matrix.DenseMatrix

	paramMinDistanceDiffForMapUpdate float64
	paramMinAngleDiffForMapUpdate    float64
}

func MakeHectorSlamProcessor(mapResolution float64, mapSizeX, mapSizeY int,
	startCoords [2]float64, multiResSize int) *HectorSlamProcessor {

	hsp := new(HectorSlamProcessor)

	hsp.mapRep = maprep.MakeMapRepMultiMap(mapResolution, mapSizeX, mapSizeY, multiResSize, startCoords)
	//	hsp.mapRep = maprep.MakeMapRepSingleMap(mapResolution, mapSizeX, mapSizeY, startCoords)

	hsp.Reset()

	hsp.SetMapUpdateMinDistDiff(0.4)
	hsp.SetMapUpdateMinAngleDiff(0.9)

	// Initialize matrix
	hsp.lastScanMatchCov = matrix.Eye(3)

	return hsp
}

func MakeHectorSlamProcessorFromMapRep(mapRep maprep.MapRepresentation) *HectorSlamProcessor {

	hsp := new(HectorSlamProcessor)

	hsp.mapRep = mapRep

	// hsp.Reset()

	hsp.SetMapUpdateMinDistDiff(0.4)
	hsp.SetMapUpdateMinAngleDiff(0.9)

	// Initialize matrix
	hsp.lastScanMatchCov = matrix.Eye(3)

	return hsp
}

func (hsp *HectorSlamProcessor) Update(dataContainer *datacontainer.DataContainer,
	poseHintWorld [3]float64) {

	newPoseEstimateWorld := hsp.mapRep.MatchData(poseHintWorld, dataContainer, hsp.lastScanMatchCov)

	hsp.lastScanMatchPose = newPoseEstimateWorld

	if utils.PoseDifferenceLargerThan(newPoseEstimateWorld, hsp.lastMapUpdatePose, hsp.paramMinDistanceDiffForMapUpdate, hsp.paramMinAngleDiffForMapUpdate) {
		hsp.mapRep.UpdateByScan(dataContainer, newPoseEstimateWorld)

		hsp.mapRep.OnMapUpdated()
		hsp.lastMapUpdatePose = newPoseEstimateWorld
	}
}

func (hsp *HectorSlamProcessor) Reset() {
	hsp.lastMapUpdatePose = [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	hsp.lastScanMatchPose = [3]float64{}

	hsp.mapRep.Reset()
}

func (hsp *HectorSlamProcessor) GetLastScanMatchPose() [3]float64 {
	return hsp.lastScanMatchPose
}

func (hsp *HectorSlamProcessor) GetLastScanMatchCovariance() *matrix.DenseMatrix {
	return hsp.lastScanMatchCov
}

func (hsp *HectorSlamProcessor) GetLastMapUpdatePose() [3]float64 {
	return hsp.lastMapUpdatePose
}

func (hsp *HectorSlamProcessor) GetScaleToMap() float64 {
	return hsp.mapRep.GetScaleToMap()
}

func (hsp *HectorSlamProcessor) GetMapLevels() int {
	return hsp.mapRep.GetMapLevels()
}

func (hsp *HectorSlamProcessor) GetGridMap() gridmap.OccGridMap {
	return hsp.mapRep.GetGridMap(0)
}

func (hsp *HectorSlamProcessor) GetGridMapByLevel(level int) gridmap.OccGridMap {
	return hsp.mapRep.GetGridMap(level)
}

func (hsp *HectorSlamProcessor) AddMapMutex(i int, mapMutex *sync.RWMutex) {
	hsp.mapRep.AddMapMutex(i, mapMutex)
}

func (hsp *HectorSlamProcessor) GetMapMutex(i int) *sync.RWMutex {
	return hsp.mapRep.GetMapMutex(i)
}

func (hsp *HectorSlamProcessor) SetUpdateFactorFree(freeFactor float64) {
	hsp.mapRep.SetUpdateFactorFree(freeFactor)
}

func (hsp *HectorSlamProcessor) SetUpdateFactorOccupied(occupiedFactor float64) {
	hsp.mapRep.SetUpdateFactorOccupied(occupiedFactor)
}

func (hsp *HectorSlamProcessor) SetMapUpdateMinDistDiff(minDist float64) {
	hsp.paramMinDistanceDiffForMapUpdate = minDist
}

func (hsp *HectorSlamProcessor) SetMapUpdateMinAngleDiff(minAngle float64) {
	hsp.paramMinAngleDiffForMapUpdate = minAngle
}

func (hsp *HectorSlamProcessor) GetMapRepresentation() maprep.MapRepresentation {
	return hsp.mapRep
}
