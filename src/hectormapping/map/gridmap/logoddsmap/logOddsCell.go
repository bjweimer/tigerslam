package logoddsmap

import (
	"bytes"
	"encoding/gob"

	"hectormapping/map/gridmap"
)

// Provides log odds of occupancy probability representation for cells in an
// occupancy grid map.
type LogOddsCell struct {
	logOddsVal  float64
	updateIndex int
}

// Sets the cell value to val
func (loc *LogOddsCell) Set(val float64) {
	loc.logOddsVal = val
}

// Returns the value of the cell
func (loc *LogOddsCell) GetValue() float64 {
	return loc.logOddsVal
}

// Returns whether the cell is occupied
func (loc *LogOddsCell) IsOccupied() bool {
	return loc.logOddsVal > 0.0
}

//  whether the cell is free
func (loc *LogOddsCell) IsFree() bool {
	return loc.logOddsVal < 0.0
}

// Reset cell to prior probability
func (loc *LogOddsCell) ResetGridCell() {
	loc.logOddsVal = 0.0
	loc.updateIndex = -1
}

// Copy the cell
func (loc *LogOddsCell) Copy() gridmap.Cell {
	return &LogOddsCell{
		logOddsVal:  loc.logOddsVal,
		updateIndex: loc.updateIndex,
	}
}

// Return the update index
func (loc *LogOddsCell) GetUpdateIndex() int {
	return loc.updateIndex
}

// Set the update index
func (loc *LogOddsCell) SetUpdateIndex(index int) {
	loc.updateIndex = index
}

// Gob encode the cell
func (loc *LogOddsCell) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	// Encode logOddsVal
	err := encoder.Encode(loc.logOddsVal)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Gob decode the cell
func (loc *LogOddsCell) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	err := decoder.Decode(&loc.logOddsVal)
	if err != nil {
		return err
	}

	return nil
}
