// Package gridmapimages makes images from Hector Mapping maps, such as
// full images and tiles.
package mapimages

import (
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"

	"robot/tools/intmath"

	"hectormapping/map/gridmap"
	"hectormapping/map/maprep"
)

const TILE_SIZE = 256

var freeColor = color.Gray{255}
var occupiedColor = color.Gray{0}
var unknownColor = color.Gray{128}

// GetMapImage produces a 1:1 Image of the map representation
func GetMapImage(mapRep maprep.MapRepresentation) (image.Image, error) {
	gridMap := mapRep.GetGridMap(0)

	return GetGridmapImage(gridMap)
}

func GetGridmapImage(gridMap gridmap.OccGridMap) (image.Image, error) {
	dims := gridMap.GetMapDimensions()

	im := image.NewGray(image.Rect(0, 0, dims[0], dims[1]))

	// Produce the map using several Go routines, splitting the job into nprocs
	// jobs, each running fillInSubImage. The image is split into nprocs rows,
	// each consisting of sizeY/nprocs (approximately) lines.
	nProcs := runtime.NumCPU()
	rowSize := int(math.Ceil(float64(dims[1]) / float64(nProcs)))
	var wg sync.WaitGroup
	for i := 0; i < nProcs; i++ {
		wg.Add(1)
		go func(rowNumber int) {
			rect := image.Rect(0, (nProcs-1-rowNumber)*rowSize, gridMap.GetSizeX(), (nProcs-rowNumber)*rowSize)
			subImage := im.SubImage(rect)
			fillInSubImage(subImage, gridMap, 0, rowNumber*rowSize, 1.0)
			wg.Done()

		}(i)
	}

	// Wait for the go routines to complete
	wg.Wait()

	return im, nil
}

// GetMapTile produces a tile of a TILE_SIZE size, at zoomLevel. The tile is
// chosen by tileX and tileY in tile coordinates.
func GetMapTile(mapRep maprep.MapRepresentation, zoomLevel uint,
	tileX, tileY int) (image.Image, error) {

	// Create the image
	im := image.NewGray(image.Rect(0, 0, TILE_SIZE, TILE_SIZE))

	// Get the number of tiles in each direction. This is determined by the
	// zoomLevel alone. zoomLevel 0 means there is 1 tile, 1 means 2 tiles in
	// each direction, 2 means 4 tiles in each direction, and so on.
	numTiles := (1 << zoomLevel)

	// Choose the map to use for this zoomLevel
	var gridMap gridmap.OccGridMap
	totalPixels := numTiles * TILE_SIZE
	for i := mapRep.GetMapLevels() - 1; i >= 0; i-- {
		if intmath.Max(mapRep.GetGridMap(i).GetSizeX(), mapRep.GetGridMap(i).GetSizeY()) >= totalPixels || i == 0 {
			gridMap = mapRep.GetGridMap(i)
			break
		}
	}
	gridMapMaxSize := intmath.Max(gridMap.GetSizeX(), gridMap.GetSizeY())

	// Determine the number of cells in the map per tile
	cellsPerTile := gridMapMaxSize / numTiles

	// Get the start and end positions of the tile in the map
	startX := tileX * cellsPerTile
	startY := (numTiles - 1 - tileY) * cellsPerTile
	endX := startX + cellsPerTile
	endY := startY + cellsPerTile

	// If we're outside the map, return the blank image
	if !gridMap.HasGridValue(startX, startY) && !gridMap.HasGridValue(endX, endY) {
		return im, nil
	}

	// Get the step size, i.e. how much to increment in x and y direction for
	// each step -- might not be integer. Should be the same in both X and Y
	// direction.
	stepSize := float64(cellsPerTile) / float64(TILE_SIZE)

	// Use nProcs Go routines to fill in the image (split into nProcs rows)
	var waitGroup sync.WaitGroup
	nProcs := runtime.NumCPU()
	rowHeight := TILE_SIZE / nProcs
	for i := 0; i < nProcs; i++ {
		waitGroup.Add(1)
		go func(rowNumber int) {
			subImage := im.SubImage(image.Rect(0, (nProcs-1-rowNumber)*rowHeight, TILE_SIZE, ((nProcs-1-rowNumber)+1)*rowHeight))
			fillInSubImage(subImage, gridMap, startX, startY+int(float64(rowNumber*rowHeight)*stepSize), stepSize)
			waitGroup.Done()
		}(i)
	}

	waitGroup.Wait()

	return im, nil
}

// Fill in a sub-image im, based on gridMap, with start coordinates startX and
// startY (integer map coordinates), with stepSize determining how many cells
// are in a pixel.
func fillInSubImage(rawimg image.Image, gridMap gridmap.OccGridMap, startX, startY int, stepSize float64) {
	xCoord := float64(startX) + 0.5
	yCoord := float64(startY) + 0.5

	im := rawimg.(*image.Gray)

	// Loop through the pixels of the subImage
	xMin, xMax := im.Bounds().Min.X, im.Bounds().Max.X
	yMin, yMax := im.Bounds().Min.Y, im.Bounds().Max.Y
	for i := xMin; i < xMax; i++ {
		for j := yMax - 1; j >= yMin; j-- {

			// Get the integer map coordinates
			x := int(xCoord)
			y := int(yCoord)

			// If the cell doesn't have any value, skip
			if !gridMap.HasGridValue(x, y) {
				continue
			}

			// Obtain the cell, fill in the image
			cell := gridMap.GetCell(x, y)
			if cell.IsFree() {
				im.Set(i, j, freeColor)
			} else if cell.IsOccupied() {
				im.Set(i, j, occupiedColor)
			} else {
				im.Set(i, j, unknownColor)
			}

			yCoord += stepSize

		}
		xCoord += stepSize
		yCoord = float64(startY)
	}
}
