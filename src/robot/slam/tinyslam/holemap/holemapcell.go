package holemap

import (
	"bytes"
	"encoding/gob"
	"hectormapping/map/gridmap"
)

type HoleMapCell struct {
	val         uint16
	updateIndex int
}

const NO_OBSTACLE = 65500
const OBSTACLE = 0

// ResetGridCell()
// Copy() Cell
// GetValue() float64
// IsFree() bool
// IsOccupied() bool
// GetUpdateIndex() int
// SetUpdateIndex(int)

// gob.GobEncoder
// gob.GobDecoder

func (hmc *HoleMapCell) ResetGridCell() {
	hmc.val = 65500 / 2
}

func (hmc *HoleMapCell) Copy() gridmap.Cell {
	return &HoleMapCell{
		val: hmc.val,
	}
}

func (hmc *HoleMapCell) GetValue() float64 {
	return float64(hmc.val)
}

func (hmc *HoleMapCell) Set(val float64) {
	hmc.val = uint16(val)
}

func (hmc *HoleMapCell) IsFree() bool {
	return hmc.val > 65500/2
}

func (hmc *HoleMapCell) IsOccupied() bool {
	return hmc.val < 65500/2
}

func (hmc *HoleMapCell) GetUpdateIndex() int {
	return hmc.updateIndex
}

func (hmc *HoleMapCell) SetUpdateIndex(index int) {
	hmc.updateIndex = index
}

func (hmc *HoleMapCell) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	err := encoder.Encode(hmc.val)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (hmc *HoleMapCell) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	err := decoder.Decode(&hmc.val)
	if err != nil {
		return err
	}

	return nil
}
