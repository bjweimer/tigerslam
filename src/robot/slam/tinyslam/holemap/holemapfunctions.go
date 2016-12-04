package holemap

import (
	"hectormapping/map/gridmap"
)

// // Provides functions related to the updating of a gridmap of a specific cell
// // representation.
// type GridFunctions interface {
// 	UpdateSetOccupied(Cell)
// 	UpdateSetFree(Cell)
// 	UpdateUnsetFree(Cell)
// 	GetGridProbability(Cell) float64
// 	SetUpdateOccupiedFactor(float64)
// 	SetUpdateFreeFactor(float64)
// }

type HoleMapFunctions struct {
	//?
}

func MakeHoleMapFunctions() *HoleMapFunctions {
	f := new(HoleMapFunctions)

	return f
}

func (f *HoleMapFunctions) ConvertToHoleMapCell(cell gridmap.Cell) *HoleMapCell {
	return cell.(*HoleMapCell)
}

func (f *HoleMapFunctions) GetGridProbability(cell gridmap.Cell) float64 {
	return 0.0
}
