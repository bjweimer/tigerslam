package cache

import (

)

// Implements CacheMethod. Caches filtered grid map accesses in a an array of
// the same size as the map. Each cached element stores a value and a cache
// index. The cache index is incremented to "flush" the cache, done by the
// Reset() method, so that values with different cache indeces than the current
// are known to be invalid.
type GridMapCacheArray struct {
	
	// Array used for caching data
	cacheArray []*CachedMapElement
	
	// The cache iteration index value
	currCacheIndex int
	
	// The size of the array
	arrayDimensions [2]int
}

type CachedMapElement struct {
	val float64
	index int
}

func MakeCachedMapElement() *CachedMapElement {
	return &CachedMapElement{
		index: -1,
	}
}

func MakeGridMapCacheArray() *GridMapCacheArray {
	return &GridMapCacheArray{
		arrayDimensions: [2]int{-1, -1},
		currCacheIndex: 0,
	}
}

// Resets/deletes the cached data
func (gmca *GridMapCacheArray) ResetCache() {
	gmca.currCacheIndex++
}

// Checks whether cached data for coords are available. If this is the case,
// writes data into val.
// @param index The index
// @param val Reference to the float the data is written to if available
// @return Indicates if cached data is available
func (gmca *GridMapCacheArray) ContainsCachedData(index int, val *float64) bool {
	elem := gmca.cacheArray[index]
	
	if elem.index == gmca.currCacheIndex {
		*val = elem.val
		return true
	}
	
	return false
}

// Caches float value val for the given index.
// @param index The index
// @param val The value to be cached for coordinates
func (gmca *GridMapCacheArray) CacheData(index int, val float64) {
	gmca.cacheArray[index].index = gmca.currCacheIndex
	gmca.cacheArray[index].val = val
}

// Sets the map size and resizes the cache array accordingly
// @param sizeIn The map size
func (gmca *GridMapCacheArray) SetMapSize(newDimensions [2]int) {
	gmca.SetArraySize(newDimensions)
}

// Creates a cache array of size sizeIn
// @param sizeIn The size of the array (in two dimensions)
func (gmca *GridMapCacheArray) createCacheArray(newDimensions [2]int) {
	size := newDimensions[0] * newDimensions[1]
	gmca.cacheArray = make([]*CachedMapElement, size)
	
	for i := range gmca.cacheArray {
		gmca.cacheArray[i] = MakeCachedMapElement()
	}
}

// Deletes the existing cache array
func (gmca *GridMapCacheArray) deleteCacheArray() {
	gmca.cacheArray = nil
}

// Sets a new cache array size
func (gmca *GridMapCacheArray) SetArraySize(newDimensions [2]int) {
	if gmca.arrayDimensions != newDimensions {
		if gmca.cacheArray != nil {
			gmca.deleteCacheArray()
		}
		gmca.createCacheArray(newDimensions)
	}
}