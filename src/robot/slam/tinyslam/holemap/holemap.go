// Package holemap implements maps for tinySLAM. They are described in the
// tinySLAM article, and resembles likelihood fields described in Probabilistic
// Robotics (book). Each cell is represented by a floating point value, and
// its value can be thought of as representing an estimate of the distance to
// the nearest obstacle. This works by ray casting. When a LIDAR measurement is
// painted on the map, values are lower in the pixels around the point where
// the LASER ray stopped (and rising again for some pixels after). This creates
// a "hole" around cells where the LASER rays have stopped, as described in the
// tinySLAM article.
package holemap

import (
	"bytes"
	"encoding/gob"

	"hectormapping/map/gridmap/mapdimensionproperties"
	"hectormapping/map/gridmap/occbase"
)

type HoleMap struct {
	occbase.OccGridMapBase
}

func MakeHoleMap(mapResolution float64, size [2]int, offset [2]float64) *HoleMap {
	m := &HoleMap{
		OccGridMapBase: *occbase.MakeOccGridMapBase(mapResolution, size, offset, &HoleMapCell{}),
	}
	m.OccGridMapBase.ConcreteGridFunctions = MakeHoleMapFunctions()

	return m
}

func (hm *HoleMap) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	var cells []*HoleMapCell
	err := decoder.Decode(&cells)
	if err != nil {
		return err
	}

	var mdp mapdimensionproperties.MapDimensionProperties
	err = decoder.Decode(&mdp)
	if err != nil {
		return err
	}

	hm.SetCellExample(&HoleMapCell{})
	hm.SetDimensionProperties(&mdp)

	hm.ConcreteGridFunctions = MakeHoleMapFunctions()

	// Set the cells
	for i := range cells {
		hm.SetCell(cells[i], i)
	}

	return nil
}
