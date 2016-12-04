// Package binarymap implements maps where the cells are either 0 or 1, that
// is, either obstacle or free.
package binarymap

import (
	"bytes"
	"encoding/gob"
	"math"

	"hectormapping/map/gridmap"
	"hectormapping/map/gridmap/mapdimensionproperties"
	"hectormapping/map/gridmap/occbase"

	"robot/tools/intmath"
)

// OccGridMapBinary is a OccGridMap which returns BinaryCells
type OccGridMapBinary struct {
	occbase.OccGridMapBase
}

func MakeOccGridMapBinary(mapResolution float64, size [2]int, offset [2]float64) *OccGridMapBinary {
	m := &OccGridMapBinary{
		OccGridMapBase: *occbase.MakeOccGridMapBase(mapResolution, size, offset, &BinaryCell{}),
	}
	m.OccGridMapBase.ConcreteGridFunctions = MakeGridMapBinaryFunctions()

	return m
}

// Takes in an OccGridMap, returns an empty binary map shrunken by a factor
// shrinkFactor.
func BinaryMapFromOccGridMap(occMap gridmap.OccGridMap, shrinkFactor int) *OccGridMapBinary {
	return shrunkenBinaryMap(occMap, shrinkFactor)
}

// Produce shrunken MapDimensionProperties
func shrunkenBinaryMap(original gridmap.OccGridMap, shrinkFactor int) *OccGridMapBinary {
	fShrinkFactor := float64(shrinkFactor)

	mapResolution := original.GetCellLength() * fShrinkFactor

	offset := original.GetMapDimProperties().GetTopLeftOffset()
	//	offset[0] /= fShrinkFactor
	//	offset[1] /= fShrinkFactor

	sizeX := original.GetSizeX() / shrinkFactor
	sizeY := original.GetSizeY() / shrinkFactor

	return MakeOccGridMapBinary(mapResolution, [2]int{sizeX, sizeY}, offset)
}

func (o *OccGridMapBinary) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	var cells []*BinaryCell
	err := decoder.Decode(&cells)
	if err != nil {
		return err
	}

	var mdp mapdimensionproperties.MapDimensionProperties
	err = decoder.Decode(&mdp)
	if err != nil {
		return err
	}

	o.SetCellExample(&BinaryCell{})
	o.SetDimensionProperties(&mdp)

	o.ConcreteGridFunctions = MakeGridMapBinaryFunctions()

	// Set the cells
	for i := range cells {
		o.GetCellByIndex(i).Set(cells[i].GetValue())
	}

	return nil
}

// Make a shrunken representation of a occMap and fill in the new values.
// The cell radius (in meters) is added in each direction to the area checked
// for occupancy for each cell in the binary map. In other words, if the
// OccMap (original map) has any occupied cells within the corresponding cell
// in the binary map + radius, the binary cell is set occupied.
func MakeShrunkenBinaryMap(occMap gridmap.OccGridMap, shrinkFactor int, checkRadius float64) *OccGridMapBinary {
	binMap := BinaryMapFromOccGridMap(occMap, shrinkFactor)
	binMap.ConcreteGridFunctions = MakeGridMapBinaryFunctions()

	// Radius in number of cells
	cellRadius := int(math.Ceil(checkRadius / occMap.GetCellLength()))

	// Loop through the new binary map. For each cell, loop through the
	// corresponding cells in the original map, and check for occupied cells.
	for i := 0; i < binMap.GetSizeX(); i++ {
		for j := 0; j < binMap.GetSizeY(); j++ {

			if pieceIsFree(occMap, i*shrinkFactor-cellRadius, j*shrinkFactor-cellRadius, shrinkFactor+2*cellRadius) {
				binMap.ConcreteGridFunctions.UpdateSetFree(binMap.GetCell(i, j))
			}

		}
	}

	return binMap
}

// Check a piece of size size^2 for occupied pixels, starting at (xMin, yMin)
// for occupied cells. Return true if there are no occupied cells in the piece.
func pieceIsFree(occMap gridmap.OccGridMap, xMin, yMin, size int) bool {

	xMin = intmath.Max(xMin, 0)
	yMin = intmath.Max(yMin, 0)

	xMax := intmath.Min(xMin+size, occMap.GetSizeX())
	yMax := intmath.Min(yMin+size, occMap.GetSizeY())

	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			if occMap.GetCell(x, y).IsOccupied() {
				return false
			}
		}
	}

	return true
}
