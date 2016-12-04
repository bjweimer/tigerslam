package gridmap

import (
    "testing"
    "image/png"
    "os"
    "runtime"
)

func MakeSmileySimpleMap() *SimpleMap {
	gridmap := MakeSimpleMap(1024, 100)
	gridmap.Fill(SimpleCell(32000))
	
	// Paint a smiley
	white := SimpleCell(65535)		//  x x
	gridmap.Set(201, 200, white)	//
	gridmap.Set(203, 200, white)	// x   x
	gridmap.Set(200, 202, white)	//  xxx
	gridmap.Set(201, 203, white)
	gridmap.Set(202, 203, white)
	gridmap.Set(203, 203, white)
	gridmap.Set(204, 202, white)
	
	return gridmap
}

func TestMakeSimpleMap(t *testing.T) {
	var gridmap GridMap
	
	gridmap = MakeSimpleMap(1000, 100)
	
	t.Logf("Size of gridmap is %d^2 cells, %d^2 meters", gridmap.Size(), gridmap.SizeMeters())
}

func TestSimpleMapSetAt(t *testing.T) {
	gridmap := MakeSimpleMap(1000, 100)
	
	cell := SimpleCell(1234)
	
	gridmap.Set(500, 500, cell)
	
	if gridmap.At(500, 500) != cell {
		t.Errorf("Gridmap(500,500) was %d, should have been %d", gridmap.At(500, 500), cell)
	} else {
		t.Logf("Gridmap(500, 500) is %d", gridmap.At(500, 500))
	}
}

func TestSimpleMapFill(t *testing.T) {
	gridmap := MakeSimpleMap(1000, 100)
	cell := SimpleCell(65535)
	gridmap.Fill(cell)
	
	if gridmap.At(724, 135) != cell {
		t.Errorf("Gridmap(724,135) was %d, should have been %d", gridmap.At(724, 135), cell)
	}
}

func TestSimpleImage(t *testing.T) {
	
	gridmap := MakeSmileySimpleMap()
		
	image := gridmap.Image(4, 3, 3)
	if image.Bounds().Dx() != 64 {
		t.Errorf("Image size was %d * %d, should have been 64 * 64", image.Bounds().Dx(), image.Bounds().Dy())
	} else {
		t.Logf("Image size: %d * %d", image.Bounds().Dx(), image.Bounds().Dy())
	}
	
	file, err := os.Create("testoutput/simpleimagezoom.png")
	if err != nil {
		t.Error(err)
	}
	
	png.Encode(file, image)
}

func TestSimpleFullImage(t *testing.T) {
	gridmap := MakeSmileySimpleMap()
	
	image := gridmap.Image(0, 0, 0)
	
	file, err := os.Create("testoutput/simpleimage.png")
	if err != nil {
		t.Error(err)
	}
	
	png.Encode(file, image)
}

func TestSimpleImageTileGrid(t *testing.T) {
	gridmap := MakeSimpleMap(2048, 100)
	gridmap.Fill(SimpleCell(32000))
	
	// Draw a grid
	for i := 0; i < 2048; i++ {
		for j := 0; j < 2048; j++ {
			if i == 1024 || j == 1024 {
				gridmap.Set(i, j, SimpleCell(65000))
			}
		}
	}
	
	file, err := os.Create("testoutput/SimpleImageTileGrid.png")
	if err != nil {
		t.Error(err)
	}
	
	image, err := gridmap.ImageTile(3, 4, 4)
	if err != nil {
		t.Error(err)
	}
	png.Encode(file, image)
}

func BenchmarkSimpleImage1024(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())
	gridmap := MakeSimpleMap(1024, 100)
	b.Logf("Tilesize: %d", gridmap.Image(4, 0, 0).Bounds().Dx())
	b.Logf("Mapsize in meters: %f", gridmap.SizeMeters())
	b.StartTimer()
	
	for i := 0; i < b.N; i++ {
		gridmap.Image(4, 0, 0)
	}
}

func BenchmarkSimpleImage8192(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())
	gridmap := MakeSimpleMap(8192, 100)
	b.Logf("Tilesize: %d", gridmap.Image(7, 0, 0).Bounds().Dx())
	b.Logf("Mapsize in meters: %f", gridmap.SizeMeters())
	b.StartTimer()
	
	for i := 0; i < b.N; i++ {
		gridmap.Image(7, 0, 0)
	}
}

func BenchmarkSimpleImage32768(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())
	gridmap := MakeSimpleMap(32768, 100)
	b.Logf("Tilesize: %d", gridmap.Image(9, 0, 0).Bounds().Dx())
	b.Logf("Mapsize in meters: %f", gridmap.SizeMeters())
	b.StartTimer()
	
	for i := 0; i < b.N; i++ {
		gridmap.Image(9, 0, 0)
	}
}