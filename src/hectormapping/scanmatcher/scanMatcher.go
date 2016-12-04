package scanmatcher

import (
	"github.com/skelterjohn/go.matrix"
	
	"hectormapping/datacontainer"
	"hectormapping/map/gridmap/occbase"
	
	"hectormapping/utils"
)

type ScanMatcher struct {
	dTr *matrix.DenseMatrix
	H   *matrix.DenseMatrix
	// Drawing Interface
	// Debug Info Interface
}

func MakeScanMatcher() *ScanMatcher {
	return &ScanMatcher{
		H: matrix.Zeros(3, 3),
		dTr: matrix.Zeros(3, 1),
	}
}

func (sm *ScanMatcher) MatchData(beginEstimateWorld [3]float64,
	gridMapUtil *occbase.OccGridMapUtil, dataContainer *datacontainer.DataContainer,
	covMatrix *matrix.DenseMatrix, maxIterations int) [3]float64 {

	// If drawInterface ...

	if dataContainer.GetSize() == 0 {
		return beginEstimateWorld
	}
	
	beginEstimateMap := gridMapUtil.GetMapCoordsPose(beginEstimateWorld)
	estimate := beginEstimateMap
	
	sm.EstimateTransformationLogLh(&estimate, gridMapUtil, dataContainer)
	
	numIter := maxIterations
	
	for i := 0; i < numIter; i++ {
		sm.EstimateTransformationLogLh(&estimate, gridMapUtil, dataContainer)
	}
	
	estimate[2] = utils.NormalizeAngle(estimate[2])

	*covMatrix = *sm.H.Copy()
	
	return gridMapUtil.GetWorldCoordsPose(estimate)

}

func (sm *ScanMatcher) EstimateTransformationLogLh(estimate *[3]float64,
		gridMapUtil *occbase.OccGridMapUtil, dataPoints *datacontainer.DataContainer) bool {
	
	gridMapUtil.GetCompleteHessianDerivs(estimate, dataPoints, sm.H, sm.dTr)
	
	if (sm.H.Get(0, 0) != 0.0) && (sm.H.Get(1, 1) != 0.0) {
		
		Hinv, err := sm.H.Inverse()
		if err != nil {
			panic(err)
		}
		
		tmpMatrix, err := Hinv.TimesDense(sm.dTr)
		if err != nil {
			panic(err)
		}
		searchDir := [3]float64{tmpMatrix.Get(0, 0), tmpMatrix.Get(1, 0), tmpMatrix.Get(2, 0)}
		
		if searchDir[2] > 0.2 {
			searchDir[2] = 0.2
//			logger.Println("SearchDir angle change too large")
		} else if searchDir[2] < -0.2 {
			searchDir[2] = -0.2
//			logger.Println("SearchDir angle change too large")
		}
		
		estimate = sm.UpdateEstimatedPose(estimate, searchDir)
		return true
		
	}
	
	return false
}

func (sm *ScanMatcher) UpdateEstimatedPose(estimate *[3]float64, change [3]float64) *[3]float64 {
	estimate[0] += change[0]
	estimate[1] += change[1]
	estimate[2] += change[2]
//	for i := range estimate {
//		estimate[i] += change[i]
//	}
	return estimate
}
