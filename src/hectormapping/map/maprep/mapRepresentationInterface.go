package maprep

import (
	"sync"

	"github.com/skelterjohn/go.matrix"
	
	"hectormapping/map/gridmap"
	"hectormapping/datacontainer"
)

type MapRepresentation interface {
	Reset()
	GetScaleToMap() float64
	GetMapLevels() int
	GetGridMap(mapLevel int) gridmap.OccGridMap
	AddMapMutex(i int, mapMutex *sync.RWMutex)
	GetMapMutex(i int) *sync.RWMutex
	OnMapUpdated()
	MatchData(beginEstimateWorld [3]float64, dataContainer *datacontainer.DataContainer, covMatrix *matrix.DenseMatrix) [3]float64
	UpdateByScan(dataContainer *datacontainer.DataContainer, robotPoseWorld [3]float64)
	SetUpdateFactorFree(freeFactor float64)
	SetUpdateFactorOccupied(occupiedFactor float64)
}