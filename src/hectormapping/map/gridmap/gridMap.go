package gridmap

import (
	"encoding/gob"

	matrix "github.com/skelterjohn/go.matrix"

	mdp "hectormapping/map/gridmap/mapdimensionproperties"
)

type GridMap interface {
	HasGridValue(x, y int) bool
	GetMapDimensions() [2]int
	GetSizeX() int
	GetSizeY() int
	PointOutOfMapBounds(pointMapCoords [2]float64) bool
	Reset()
	Clear()
	GetMapDimProperties() *mdp.MapDimensionProperties
	GetCell(x, y int) Cell
	GetCellByIndex(index int) Cell
	SetMapGridSize(newMapDims [2]int)
	GetWorldCoords(mapCoords [2]float64) [2]float64
	GetMapCoords(worldCoords [2]float64) [2]float64
	GetWorldCoordsPose(mapPose [3]float64) [3]float64
	GetMapCoordsPose(worldPose [3]float64) [3]float64
	GetScaleToMap() float64
	GetCellLength() float64
	GetWorldTmap() *matrix.DenseMatrix
	GetWorldTmap3D() *matrix.DenseMatrix
	GetMapTworld() *matrix.DenseMatrix
	SetUpdated()
	GetUpdateIndex() int
	GetMapExtends(xMax, yMax, xMin, yMin *int) bool

	gob.GobEncoder
	gob.GobDecoder
}

// The Cell interface enables the map to use different cell representations.
type Cell interface {
	ResetGridCell()
	Copy() Cell
	GetValue() float64
	IsFree() bool
	IsOccupied() bool
	GetUpdateIndex() int
	SetUpdateIndex(int)
	Set(float64)

	gob.GobEncoder
	gob.GobDecoder
}

// Provides functions related to the updating of a gridmap of a specific cell
// representation.
type GridFunctions interface {
	UpdateSetOccupied(Cell)
	UpdateSetFree(Cell)
	UpdateUnsetFree(Cell)
	GetGridProbability(Cell) float64
	SetUpdateOccupiedFactor(float64)
	SetUpdateFreeFactor(float64)
}
