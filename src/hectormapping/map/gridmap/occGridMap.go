package gridmap

import (
	"hectormapping/datacontainer"
)

type OccGridMap interface {
	GridMap
	UpdateSetOccupied(index int)
	UpdateSetFree(index int)
	UpdateUnsetFree(index int)
	GetGridProbabilityMap(xMap, yMap int) float64
	GetGridProbabilityMapByIndex(index int) float64
	IsOccupied(xMap, yMap int) bool
	IsFree(xMap, yMap int) bool
	IsOccupiedByIndex(index int) bool
	IsFreeByIndex(index int) bool
	GetObstacleThreshold() float64
	SetUpdateFreeFactor(factor float64)
	SetUpdateOccupiedFactor(factor float64)
	UpdateByScan(dataContainer *datacontainer.DataContainer, robotPoseWorld [3]float64)
}