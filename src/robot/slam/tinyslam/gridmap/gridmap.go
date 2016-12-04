package gridmap

import (
	"errors"
	"image"
)

// GridMap specifies an interface for all GridMaps
type GridMap interface {
	At(x, y int) Cell
	Set(x, y int, cell Cell) error
	SizeMeters() float64
	Size() int
	Fill(cell Cell) error
	Image(zoomLevel uint, tileX, tileY int) image.Image
	WorldToCellCoordinate(world float64) float64
}

// GenericGridMap holds data common for all GridMaps
type GenericGridMap struct {
	size int			// Size in cells
	resolution int		// Cells per meter
}

// Calculate the size in meters
func (ggm *GenericGridMap) SizeMeters() float64 {
	return float64(ggm.size) / float64(ggm.resolution)
}

// Size in number of cells
func (ggm *GenericGridMap) Size() int {
	return ggm.size
}

func (ggm *GenericGridMap) WorldToMapCoordinate(world float64) float64 {
	return world / ggm.SizeMeters() * float64(ggm.Size())
}

// SimpleMap is a map of uint16 equivalent values. Typically used to model
// opacity of a map, where 2^16 is used to model free space and 0 is occupied
// space.
type SimpleMap struct {
	GenericGridMap
	Cells []SimpleCell
}

// Construct a SimpleMap
func MakeSimpleMap(size, resolution int) (*SimpleMap) {
	buf := make([]SimpleCell, size * size)
	
	return &SimpleMap{
		GenericGridMap{
			size,
			resolution,
		},
		buf,
	}
}

// Get the value of a specific coordinate.
func (im *SimpleMap) At(x, y int) Cell {
	return im.Cells[im.size * y + x]
}

// Set the value of a specific corrdinate.
func (im *SimpleMap) Set(x, y int, cell Cell) error {
	if x > im.size || y > im.size || x < 0 || y < 0 {
		return errors.New("Invalid coordinates")
	}
	
	var ok bool
	im.Cells[im.size * y + x], ok = cell.(SimpleCell)
	if !ok {
		return errors.New("Invalid cell")
	}
	
	return nil
}

// Fill map with the value of a cell.
func (im *SimpleMap) Fill(cell Cell) error {
	simpleCell, ok := cell.(SimpleCell)
	if !ok {
		return errors.New("Invalid cell")
	}
	
	for i := range im.Cells {
		im.Cells[i] = simpleCell
	}
	
	return nil
}

// Return an image at a zoomLevel of the tile (x, y). The map width and height
// is split into 2^zoomLevel tiles, e.g. zoomLevel 2 gives 4*4 = 16 tiles total.
// Tile (0,0) is in the lower left corner of the map. Zoomlevel 0 will produce
// the entire map.
func (im *SimpleMap) Image(zoomLevel uint, tileX, tileY int) image.Image {
//	tileSize := im.size / (1 << zoomLevel)
//	image := image.NewGray(image.Rect(0, 0, tileSize, tileSize))
//	
//	// Concurrent algorithm
//	// n is the number of goroutines to use
//	n := 1
//	
//	fmt.Printf("Length of cells: %d\n", len(im.cells))
//	
//	portionSize := tileSize / n
//	startChan := make(chan int, n)
//	finishedChan := make(chan bool, n)
//	
//	fmt.Printf("TileSize: %d\n", tileSize)
//	fmt.Printf("Portion size: %d\n", portionSize)
//	
//	tileStart := tileSize * (im.size * tileY + tileX)
//	deltaEnd := tileSize + im.size * tileSize / n
//	
//	// Start n goroutines
//	for p := 0; p < n; p++ {
//	
//		// Does tileSize/n lines from start, then announces its finished
//		go func() {
//			portion := <-startChan
//			start := tileStart + portion * portionSize * im.size
//			end := start + deltaEnd
//			
//			fmt.Println(end)
//			
//			for i, k, j := start, 0, portionSize*portion; i < end; i++ {
//				image.SetGray(k, j, im.cells[i].Gray())
//				
//				k++
//				if k == tileSize {
//					k = 0
//					i = i + im.size - tileSize
//					j++
//				}
//			}
//			
//			finishedChan <- true
//		}()
//		
//		// Send start value
//		startChan <- p
//	}
//	
//	// Wait for the goroutines to terminate
//	for p := 0; p < n; p++ {
//		<-finishedChan
//	}
//	
//	return image
	
	tileSize := im.size / (1 << zoomLevel)
	image := image.NewGray(image.Rect(0, 0, tileSize, tileSize))
	
	// optimized algorithm
	// i loops through the cell indeces
	// k loops through x indeces
	// j loops through y indeces
	// n is the number of goroutines to use
	
	start := tileSize * (im.size * tileY + tileX)
	end := start + im.size * tileSize
	
	for i, k, j := start, 0, 0; i < end; i++ {
		
		image.SetGray(k, j, im.Cells[i].Gray())
		
		k++
		if k == tileSize {
			k = 0
			i = i + im.size - tileSize
			j++
		}
	}
	
	return image
}

// A tile is 256*256 pixels. The zoomLevel determines how many tiles exist in
// the image. There are 2^zoomLevel tiles in the image. The tile coordinates
// determine which tile to return.
func (im *SimpleMap) ImageTile(zoomLevel uint, tileX, tileY int) (image.Image, error) {
	tileSize := 256
	
	tile := image.NewGray(image.Rect(0, 0, tileSize, tileSize))
	
	// The number of tiles in each direction
	numTiles := (1 << zoomLevel)
	if tileX < 0 || tileX > numTiles - 1 || tileY < 0 || tileY > numTiles - 1 {
		return tile, nil
	}
	
	// The number of cells covered by the tile in x and y direction
	tileNumCells := im.size / numTiles
	if tileNumCells < tileSize {
		return nil, errors.New("Zoom level too high")
	}
	
	tileStartX := tileNumCells * tileX
	tileStartY := tileNumCells * (numTiles - 1 - tileY)
	
	// A pixel is pixelSize cells in the map, in each direction
	pixelSize := tileNumCells / tileSize
	
	for pixel, row := tileStartY * im.size + tileStartX, tileSize - 1; row >= 0; pixel, row = pixel + im.size * pixelSize - tileNumCells, row - 1 {
		for col := 0; col < tileSize; pixel, col = pixel + pixelSize, col + 1 {
			tile.SetGray(col, row, im.Cells[pixel].Gray())
		}
	}
	
	return tile, nil
}