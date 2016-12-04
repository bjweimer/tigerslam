package binarymap

import (
	"hectormapping/map/gridmap"
)

// Provides functions related to a binary representation for cells in an
// occupancy grid map.
type GridMapBinaryFunctions struct {
}

func MakeGridMapBinaryFunctions() *GridMapBinaryFunctions {
	return &GridMapBinaryFunctions{}
}

func (g *GridMapBinaryFunctions) ConvertToBinaryCell(cell gridmap.Cell) *BinaryCell {
	return cell.(*BinaryCell)
}

// Update cell as occupied
func (g *GridMapBinaryFunctions) UpdateSetOccupied(cell gridmap.Cell) {
	locell := g.ConvertToBinaryCell(cell)
	locell.free = false
}

// Update cell as free
func (g *GridMapBinaryFunctions) UpdateSetFree(cell gridmap.Cell) {
	locell := g.ConvertToBinaryCell(cell)
	locell.free = true
}

// Reverse update cell as free
func (g *GridMapBinaryFunctions) UpdateUnsetFree(cell gridmap.Cell) {
	g.UpdateSetOccupied(cell)
}

// Get the probability value represented by the grid cell
func (g *GridMapBinaryFunctions) GetGridProbability(cell gridmap.Cell) float64 {
	if cell.IsFree() {
		return 0
	}
	return 1
}

func (g *GridMapBinaryFunctions) SetUpdateFreeFactor(factor float64) {
	// noop
}

func (g *GridMapBinaryFunctions) SetUpdateOccupiedFactor(factor float64) {
	// noop
}
