package mapstorage

import (
	mapdimprop "hectormapping/map/gridmap/mapdimensionproperties"
)

// MapMetaData is meta data for maps
type MapMetaData struct {
	// Name of the map (might be different from filename)
	Name string
	// Map description
	Description string
	// Store the MapDimProps here, so we can read them without having to unpack
	// the whole map
	Mdp *mapdimprop.MapDimensionProperties
	// Store whether it's a MapRepMultiMap or MapRepSingleMap
	IsMapRepSingleMap bool
	// Map type string
	MapType string

	// @TODO: Add position history, landmarks, pictures, +++
}
