package cache

import (

)

// A CacheMethod can cache a float64 for each of the cells in a map of a given
// size. The map size is specified via the SetMapSize() method.
type CacheMethod interface {

	// Clears the cache for new values
	ResetCache()
	
	// Returns true and fills in val if the values exists in the cache.
	ContainsCachedData(index int, val *float64) bool
	
	// Caches the value val at index.
	CacheData(index int, val float64)
	
	// Sets the size of the map which the cache should reflect.
	SetMapSize(dimensions [2]int)
	
}