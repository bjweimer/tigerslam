package mapdimensionproperties

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// MapDimensionProperties holds information about maps. It has methods for
// convenience.
type MapDimensionProperties struct {
	cellLength    float64
	topLeftOffset [2]float64
	mapDimensions [2]int
	mapLimitsf    [2]float64
}

func MakeMapDimensionProperties(topLeftOffsetIn [2]float64, mapDimensionsIn [2]int, cellLengthIn float64) *MapDimensionProperties {
	mdp := &MapDimensionProperties{
		topLeftOffset: topLeftOffsetIn,
		cellLength:    cellLengthIn,
	}
	mdp.SetMapCellDims(mapDimensionsIn)
	return mdp
}

// Checks to see if two MapDimensionProperties objects are equal
func (mdp *MapDimensionProperties) Equal(other *MapDimensionProperties) bool {
	return (mdp.topLeftOffset == other.topLeftOffset) &&
		(mdp.mapDimensions == other.mapDimensions) &&
		(mdp.cellLength == other.cellLength)
}

// Checks to see if two MapDimensionProperties objects have equal dimension
// properties.
func (mdp *MapDimensionProperties) HasEqualDimensionProperties(other *MapDimensionProperties) bool {
	return (mdp.mapDimensions == other.mapDimensions)
}

// Checks to see if two MapDimensionProperties objects have equal
// transformation properties (i.e. their topLeftOffsets and cellLenghts are
// equal).
func (mdp *MapDimensionProperties) HasEqualTransformationProperties(other *MapDimensionProperties) bool {
	return (mdp.topLeftOffset == other.topLeftOffset) && (mdp.cellLength == other.cellLength)
}

// Check if a point is out of map bounds
func (mdp *MapDimensionProperties) PointOutOfMapBounds(coords [2]float64) bool {
	return ((coords[0] < 0.0) || (coords[0] > mdp.mapLimitsf[0]) || (coords[1] < 0.0) || (coords[1] > mdp.mapLimitsf[1]))
}

// Set map cell dimensions
func (mdp *MapDimensionProperties) SetMapCellDims(newDims [2]int) {
	mdp.mapDimensions = newDims
	mdp.mapLimitsf[0] = float64(newDims[0]) - 2.0
	mdp.mapLimitsf[1] = float64(newDims[1]) - 2.0
}

// Set the top left offset
func (mdp *MapDimensionProperties) SetTopLeftOffset(topLeftOffsetIn [2]float64) {
	mdp.topLeftOffset = topLeftOffsetIn
}

// Set size X
func (mdp *MapDimensionProperties) SetSizeX(sX int) {
	mdp.mapDimensions[0] = sX
}

// Set size Y
func (mdp *MapDimensionProperties) SetSizeY(sY int) {
	mdp.mapDimensions[1] = sY
}

// Set cell length
func (mdp *MapDimensionProperties) SetCellLength(cl float64) {
	mdp.cellLength = cl
}

// Get top left offset
func (mdp *MapDimensionProperties) GetTopLeftOffset() [2]float64 {
	return mdp.topLeftOffset
}

// Get map dimensions, i.e. the width and height of the map in the number of
// cells.
func (mdp *MapDimensionProperties) GetMapDimensions() [2]int {
	return mdp.mapDimensions
}

// Get size X
func (mdp *MapDimensionProperties) GetSizeX() int {
	return mdp.mapDimensions[0]
}

// Get size Y
func (mdp *MapDimensionProperties) GetSizeY() int {
	return mdp.mapDimensions[1]
}

// Get cell length
func (mdp *MapDimensionProperties) GetCellLength() float64 {
	return mdp.cellLength
}

// Gob encode
func (mdp MapDimensionProperties) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	err := encoder.Encode(mdp.cellLength)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(mdp.topLeftOffset)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(mdp.mapDimensions)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(mdp.mapLimitsf)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// Gob decode
func (mdp *MapDimensionProperties) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	err := decoder.Decode(&mdp.cellLength)
	if err != nil {
		return err
	}

	err = decoder.Decode(&mdp.topLeftOffset)
	if err != nil {
		return err
	}

	err = decoder.Decode(&mdp.mapDimensions)
	if err != nil {
		return err
	}

	err = decoder.Decode(&mdp.mapLimitsf)
	if err != nil {
		return err
	}

	return nil
}

// Implements the json.Marshaler interface, i.e. returns a valid JSON []byte.
func (mdp *MapDimensionProperties) MarshalJSON() ([]byte, error) {
	data := struct {
		CellLength    float64
		TopLeftOffset [2]float64
		MapDimensions [2]int
		MapLimitsf    [2]float64
	}{
		mdp.cellLength,
		mdp.topLeftOffset,
		mdp.mapDimensions,
		mdp.mapLimitsf,
	}

	return json.Marshal(data)
}
