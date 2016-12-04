package binarymap

import (
	"bytes"
	"encoding/gob"

	"hectormapping/map/gridmap"
)

// Cell which is either occupied or free
type BinaryCell struct {
	free bool
}

// Sets the cell value
func (bc *BinaryCell) ResetGridCell() {
	bc.free = false
}

// Return a copy of the cell
func (bc *BinaryCell) Copy() gridmap.Cell {
	return &BinaryCell{
		bc.free,
	}
}

func (b *BinaryCell) Set(val float64) {
	b.free = val > 0
}

// This one must return a float64 -- maybe this should be changed so that Cells
// can return any value, but for now, use mapping -1, 1
func (bc *BinaryCell) GetValue() float64 {
	if bc.free {
		return 1
	}
	return -1
}

// Returns true if the cell is free
func (bc *BinaryCell) IsFree() bool {
	return bc.free
}

// Returns true if the cell is occupied
func (bc *BinaryCell) IsOccupied() bool {
	return !bc.free
}

func (bc *BinaryCell) GetUpdateIndex() int {
	return 0
}

func (bc *BinaryCell) SetUpdateIndex(i int) {
	// noop
}

func (b *BinaryCell) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	err := encoder.Encode(b.free)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (b *BinaryCell) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	err := decoder.Decode(&b.free)
	if err != nil {
		return err
	}

	return nil
}
