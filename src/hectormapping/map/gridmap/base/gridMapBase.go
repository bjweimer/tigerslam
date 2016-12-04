package base

import (
	"bytes"
	"encoding/gob"
	"reflect"

	"github.com/skelterjohn/go.matrix"

	"hectormapping/map/gridmap"
	mdp "hectormapping/map/gridmap/mapdimensionproperties"
)

// GridMapBase provides basic grid map functionality (creates grid, provides
// transformation from/to world coordinates). It serves as the base class for
// different map representation that may extend it's functionality.
type GridMapBase struct {
	// Map representation used with plain pointer array.
	mapArray []gridmap.Cell

	// Scaling factor from world to map
	scaleToMap float64

	// Homogenous 2D transform from map to world coordinates.
	worldTmap *matrix.DenseMatrix

	// Homogenous 3D transform from map to world coordinates.
	worldTmap3D *matrix.DenseMatrix

	// Homogenous 2D transform from world to map coordinates.
	mapTworld *matrix.DenseMatrix

	mapDimensionProperties mdp.MapDimensionProperties
	sizeX                  int
	lastUpdateIndex        int
	cellExample            gridmap.Cell
}

// Constructor, creates grid representation and transformations
func MakeGridMapBase(mapResolution float64, size [2]int, offset [2]float64, cellExample gridmap.Cell) *GridMapBase {
	gmb := &GridMapBase{
		lastUpdateIndex: -1,
		sizeX:           size[0],
		cellExample:     cellExample,
	}

	gmb.SetMapGridSize(size)
	gmb.SetMapTransformation(offset, mapResolution)

	gmb.Clear()

	return gmb
}

// Indicates if given x and y are within map bounds.
// @return True if coordinates are within map bounds.
func (gmb *GridMapBase) HasGridValue(x, y int) bool {
	return (x >= 0) && (y >= 0) && (x < gmb.GetSizeX()) && (y < gmb.GetSizeY())
}

// Get the map dimensions, i.e. the width and height of the map in the number
// of cells.
func (gmb *GridMapBase) GetMapDimensions() [2]int {
	return gmb.mapDimensionProperties.GetMapDimensions()
}

// Get size X
func (gmb *GridMapBase) GetSizeX() int {
	return gmb.mapDimensionProperties.GetSizeX()
}

// Get size Y
func (gmb *GridMapBase) GetSizeY() int {
	return gmb.mapDimensionProperties.GetSizeY()
}

// Check if a map coordinate is out of map bounds
func (gmb *GridMapBase) PointOutOfMapBounds(pointMapCoords [2]float64) bool {
	return gmb.mapDimensionProperties.PointOutOfMapBounds(pointMapCoords)
}

// Reset the map
func (gmb *GridMapBase) Reset() {
	gmb.Clear()
}

// Resets the grid cell values by using the resetGridCell() function.
func (gmb *GridMapBase) Clear() {
	size := gmb.GetSizeX() * gmb.GetSizeY()

	for i := 0; i < size; i++ {
		gmb.mapArray[i].ResetGridCell()
	}
}

// Get the map dimension properties
func (gmb *GridMapBase) GetMapDimProperties() *mdp.MapDimensionProperties {
	return &gmb.mapDimensionProperties
}

// Allocates memory for the two dimensionsl pointer array for map
// representation.
func (gmb *GridMapBase) AllocateArray(newMapDims [2]int) {
	sizeX := newMapDims[0]
	sizeY := newMapDims[1]

	gmb.mapArray = make([]gridmap.Cell, sizeX*sizeY)
	gmb.mapDimensionProperties.SetMapCellDims(newMapDims)

	// Fill array with pointers to actual cells
	cellType := reflect.TypeOf(gmb.cellExample).Elem()
	for i := range gmb.mapArray {
		gmb.mapArray[i] = reflect.New(cellType).Interface().(gridmap.Cell)
	}
}

// Delete the map array. May be basically useless in Go, but include it either
// way.
func (gmb *GridMapBase) DeleteArray() {
	if gmb.mapArray != nil {
		gmb.mapArray = nil
	}

	gmb.mapDimensionProperties.SetMapCellDims([2]int{-1, -1})
}

// Get a specific cell by coordinates
func (gmb *GridMapBase) GetCell(x, y int) gridmap.Cell {
	return gmb.mapArray[y*gmb.sizeX+x]
}

// Get a specific cell by its index in the array
func (gmb *GridMapBase) GetCellByIndex(index int) gridmap.Cell {
	return gmb.mapArray[index]
}

// Set map grid size
func (gmb *GridMapBase) SetMapGridSize(newMapDims [2]int) {
	if newMapDims != gmb.mapDimensionProperties.GetMapDimensions() {
		gmb.DeleteArray()
		gmb.AllocateArray(newMapDims)
		gmb.Reset()

		gmb.sizeX = newMapDims[0]
	}
}

// Copy constructor
func (gmb *GridMapBase) GridMapBase(other *GridMapBase) {
	gmb.AllocateArray(other.GetMapDimensions())
	*gmb = *other
}

// Assignment (=)
func (gmb *GridMapBase) Assign(other *GridMapBase) *GridMapBase {
	if !gmb.mapDimensionProperties.Equal(&other.mapDimensionProperties) {
		gmb.SetMapGridSize(other.mapDimensionProperties.GetMapDimensions())
	}

	gmb.mapDimensionProperties = other.mapDimensionProperties

	gmb.worldTmap = other.worldTmap
	gmb.mapTworld = other.mapTworld
	gmb.worldTmap3D = other.worldTmap3D

	gmb.scaleToMap = other.scaleToMap

	// @todo potential resize
	//	sizeX := gmb.GetSizeX()
	//	sizeY := gmb.GetSizeY()

	// Copy the cells -- done with memcpy in C++
	for i := range other.mapArray {
		gmb.mapArray[i] = other.mapArray[i].Copy()
	}

	return gmb
}

// Returns the world coordinates for the given map coordinates
func (gmb *GridMapBase) GetWorldCoords(mapCoords [2]float64) [2]float64 {
	worldCoords, err := gmb.worldTmap.TimesDense(matrix.MakeDenseMatrix(append(mapCoords[:], 1), 3, 1))
	if err != nil {
		panic(err)
	}
	array := worldCoords.Array()
	return [2]float64{array[0], array[1]}
}

// Returns the map coordinates for the given world coords.
func (gmb *GridMapBase) GetMapCoords(worldCoords [2]float64) [2]float64 {
	mapCoords, err := gmb.mapTworld.TimesDense(matrix.MakeDenseMatrix(append(worldCoords[:], 1), 3, 1))
	if err != nil {
		panic(err)
	}
	array := mapCoords.Array()
	return [2]float64{array[0], array[1]}
}

// Returns the world pose for the given map pose.
func (gmb *GridMapBase) GetWorldCoordsPose(mapPose [3]float64) [3]float64 {
	mapCoords := [2]float64{mapPose[0], mapPose[1]}
	worldCoords := gmb.GetWorldCoords(mapCoords)
	return [3]float64{worldCoords[0], worldCoords[1], mapPose[2]}
}

// Returns the map pose for the given world pose.
func (gmb *GridMapBase) GetMapCoordsPose(worldPose [3]float64) [3]float64 {
	worldCoords := [2]float64{worldPose[0], worldPose[1]}
	mapCoords := gmb.GetMapCoords(worldCoords)
	return [3]float64{mapCoords[0], mapCoords[1], worldPose[2]}
}

// Set dimension properties from the parameters
func (gmb *GridMapBase) SetDimensionPropertiesParameters(topLeftOffsetIn [2]float64, mapDimensionsIn [2]int, cellLengthIn float64) {
	gmb.SetDimensionProperties(mdp.MakeMapDimensionProperties(topLeftOffsetIn, mapDimensionsIn, cellLengthIn))
}

// Set dimension properties from object
func (gmb *GridMapBase) SetDimensionProperties(newMapDimProps *mdp.MapDimensionProperties) {
	// Grid map cell number has changed
	if !newMapDimProps.HasEqualDimensionProperties(&gmb.mapDimensionProperties) {
		gmb.SetMapGridSize(newMapDimProps.GetMapDimensions())
	}

	// Grid map transformation/cell size has changed
	if !newMapDimProps.HasEqualTransformationProperties(&gmb.mapDimensionProperties) {
		gmb.SetMapTransformation(newMapDimProps.GetTopLeftOffset(), newMapDimProps.GetCellLength())
	}
}

// Set the map transformations
// @param topLeftOffset The origin of the map coordinate system in world coords
// @param cellLength The cell length of the grid map
func (gmb *GridMapBase) SetMapTransformation(topLeftOffset [2]float64, cellLength float64) {
	gmb.mapDimensionProperties.SetCellLength(cellLength)
	gmb.mapDimensionProperties.SetTopLeftOffset(topLeftOffset)

	gmb.scaleToMap = 1.0 / cellLength
	var err error

	// MapTWorld should be
	// 1 0 tlo0
	// 0 1 tlo1
	// 0 0    1
	gmb.mapTworld = matrix.Eye(3)
	gmb.mapTworld.Set(0, 0, gmb.scaleToMap)
	gmb.mapTworld.Set(1, 1, gmb.scaleToMap)
	gmb.mapTworld.Set(0, 2, topLeftOffset[0]*gmb.scaleToMap)
	gmb.mapTworld.Set(1, 2, topLeftOffset[1]*gmb.scaleToMap)

	// WorldTMap3D should be the INVERSE of
	// 1 0 0 tlo0
	// 0 1 0 tlo1
	// 0 0 1 0
	// 0 0 0 1
	gmb.worldTmap3D = matrix.Eye(4)
	gmb.worldTmap3D.Set(0, 0, gmb.scaleToMap)
	gmb.worldTmap3D.Set(1, 1, gmb.scaleToMap)
	gmb.worldTmap3D.Set(0, 3, topLeftOffset[0]*gmb.scaleToMap)
	gmb.worldTmap3D.Set(1, 3, topLeftOffset[1]*gmb.scaleToMap)
	gmb.worldTmap3D, err = gmb.worldTmap3D.Inverse()
	if err != nil {
		panic(err)
	}

	// WorldTMap is the inverse of MapTWorld
	gmb.worldTmap, err = gmb.mapTworld.Inverse()
	if err != nil {
		panic(err)
	}
}

// Returns the scale factor for one unit in world coordinates to one unit in
// map coords.
func (gmb *GridMapBase) GetScaleToMap() float64 {
	return gmb.scaleToMap
}

// Returns the cell edge length of grid cells in millimeters
func (gmb *GridMapBase) GetCellLength() float64 {
	return gmb.mapDimensionProperties.GetCellLength()
}

// Returns a reference to the homogenous 2D transform from map to world
// coordinates.
func (gmb *GridMapBase) GetWorldTmap() *matrix.DenseMatrix {
	return gmb.worldTmap
}

// Returns a reference to the homogenous 3D transform from map to world
// coordinates.
func (gmb *GridMapBase) GetWorldTmap3D() *matrix.DenseMatrix {
	return gmb.worldTmap3D
}

// Returns a reference to the homogenous 2D transform from world to map
// coordinates.
func (gmb *GridMapBase) GetMapTworld() *matrix.DenseMatrix {
	return gmb.mapTworld
}

// Set updated
func (gmb *GridMapBase) SetUpdated() {
	gmb.lastUpdateIndex++
}

// Get update index
func (gmb *GridMapBase) GetUpdateIndex() int {
	return gmb.lastUpdateIndex
}

// Returns the rectangle ([xMin,yMin],[xMax,yMax]) containing non-default cell
// values.
func (gmb *GridMapBase) GetMapExtends(xMax, yMax, xMin, yMin *int) bool {
	lowerStart := -1
	upperStart := 10000

	xMaxTemp := lowerStart
	yMaxTemp := lowerStart
	xMinTemp := upperStart
	yMinTemp := upperStart

	sizeX := gmb.GetSizeX()
	sizeY := gmb.GetSizeY()

	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			if gmb.GetCell(x, y).GetValue() != 0.0 {

				if x > xMaxTemp {
					xMaxTemp = x
				}

				if x < xMinTemp {
					xMinTemp = x
				}

				if y > yMaxTemp {
					yMaxTemp = y
				}

				if y < yMinTemp {
					yMinTemp = y
				}

			}
		}
	}

	if (xMaxTemp != lowerStart) && (yMaxTemp != lowerStart) &&
		(xMinTemp != upperStart) && (yMinTemp != upperStart) {

		*xMax = xMaxTemp
		*yMax = yMaxTemp
		*xMin = xMinTemp
		*yMin = yMinTemp

		return true
	}

	return false
}

func (gmb *GridMapBase) GetCellExample() gridmap.Cell {
	return gmb.cellExample
}

func (gmb *GridMapBase) SetCell(cell gridmap.Cell, i int) {
	gmb.mapArray[i] = cell
}

func (gmb *GridMapBase) SetCellExample(cell gridmap.Cell) {
	gmb.cellExample = cell
}

// Gob encode the gridmap data so that it can be saved to file. Gob encoding has
// to be done by types which have a specific Cell type.
func (gmb *GridMapBase) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	// Encode the MapArray
	err := encoder.Encode(gmb.mapArray)
	if err != nil {
		return nil, err
	}

	// Encode MapDimensionProperties
	err = encoder.Encode(gmb.mapDimensionProperties)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
