package logoddsmap

import (
	"math"
	
	"hectormapping/map/gridmap"
)

// Provides functions related to a log odds of occupancy probability
// representation for cells in an occupancy grid map.
type GridMapLogOddsFunctions struct {
	logOddsOccupied float64
	logOddsFree float64
}

// Constructor, sets parameters like free and occupied log odds ratios.
func MakeGridMapLogOddsFunctions() *GridMapLogOddsFunctions {
	g := new(GridMapLogOddsFunctions)
	
	g.SetUpdateFreeFactor(0.4)
	g.SetUpdateOccupiedFactor(0.6)
	
	return g
}

func (g *GridMapLogOddsFunctions) ConvertToLogOddsCell(cell gridmap.Cell) *LogOddsCell {
	return cell.(*LogOddsCell)
}

// Update cell as occupied
func (g *GridMapLogOddsFunctions) UpdateSetOccupied(cell gridmap.Cell) {
	locell := g.ConvertToLogOddsCell(cell)
	if locell.logOddsVal < 50.0 {
		locell.logOddsVal += g.logOddsOccupied
	}
}

// Update cell as free
func (g *GridMapLogOddsFunctions) UpdateSetFree(cell gridmap.Cell) {
	locell := g.ConvertToLogOddsCell(cell)
	locell.logOddsVal += g.logOddsFree
}

// Reverse update cell as free
func (g *GridMapLogOddsFunctions) UpdateUnsetFree(cell gridmap.Cell) {
	locell := g.ConvertToLogOddsCell(cell)
	locell.logOddsVal -= g.logOddsFree
}

// Get the probability value represented by the grid cell.
func (g *GridMapLogOddsFunctions) GetGridProbability(cell gridmap.Cell) float64 {
	locell := g.ConvertToLogOddsCell(cell)
	odds := math.Exp(locell.logOddsVal)
	return odds / (odds + 1.0)
}

func (g *GridMapLogOddsFunctions) SetUpdateFreeFactor(factor float64) {
	g.logOddsFree = g.probToLogOdds(factor)
}

func (g *GridMapLogOddsFunctions) SetUpdateOccupiedFactor(factor float64) {
	g.logOddsOccupied = g.probToLogOdds(factor)
}

func (g *GridMapLogOddsFunctions) probToLogOdds(prob float64) float64 {
	odds := prob / (1.0 - prob)
	return math.Log(odds)
}