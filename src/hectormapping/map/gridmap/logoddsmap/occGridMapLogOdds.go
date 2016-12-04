package logoddsmap

import (
	"bytes"
	"encoding/gob"

	"hectormapping/map/gridmap/mapdimensionproperties"
	"hectormapping/map/gridmap/occbase"
)

// OccGridMapLogOdds is a OccGridMap which returns LogOddsCells
type OccGridMapLogOdds struct {
	occbase.OccGridMapBase
}

func MakeOccGridMapLogOdds(mapResolution float64, size [2]int, offset [2]float64) *OccGridMapLogOdds {
	m := &OccGridMapLogOdds{
		OccGridMapBase: *occbase.MakeOccGridMapBase(mapResolution, size, offset, &LogOddsCell{}),
	}
	m.OccGridMapBase.ConcreteGridFunctions = MakeGridMapLogOddsFunctions()

	return m
}

// Gob decode a OccGridMapLogOdds. The encoding has been done by the
// GridMapBase. For decoding, we must know what type of Cell to decode to.
func (lo *OccGridMapLogOdds) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	var cells []*LogOddsCell
	err := decoder.Decode(&cells)
	if err != nil {
		return err
	}

	var mdp mapdimensionproperties.MapDimensionProperties
	err = decoder.Decode(&mdp)
	if err != nil {
		return err
	}

	lo.SetCellExample(&LogOddsCell{})
	lo.SetDimensionProperties(&mdp)

	lo.ConcreteGridFunctions = MakeGridMapLogOddsFunctions()

	// Set the cells
	for i := range cells {
		lo.GetCellByIndex(i).Set(cells[i].GetValue())
	}

	return nil
}
