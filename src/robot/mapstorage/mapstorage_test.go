package mapstorage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image/png"
	"io"
	"os"
	"testing"

	"hectormapping/map/gridmap/logoddsmap"
	"hectormapping/map/mapimages"
	"hectormapping/map/maprep"
)

func TestSaveMultiMap(t *testing.T) {

	mapRep := maprep.MakeMapRepMultiMap(0.025, 1024, 1024, 3, [2]float64{0, 0})
	m := &Map{
		Meta: &MapMetaData{
			Name:        "Multimap",
			Description: "This is a totally awesome multilayered map.",
			MapType:     "logodds",
		},
		MapRep: mapRep,
	}

	t.Logf("Number of map levels in saved maprep: %d", m.MapRep.GetMapLevels())

	err := m.Save("testoutput\\multimap")
	if err != nil {
		t.Error(err)
	}
}

func TestLoadMultiMap(t *testing.T) {

	m, err := Load("testoutput\\multimap")
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	t.Logf("Number of map levels in loaded maprep: %d", m.MapRep.GetMapLevels())

	for i := 0; i < m.MapRep.GetMapLevels(); i++ {
		gm := m.MapRep.GetGridMap(i)
		t.Logf("Level %d has size (%d, %d).", i, gm.GetSizeX(), gm.GetSizeY())
	}
}

func TestSaveSingleMap(t *testing.T) {

	mapRep := maprep.MakeMapRepSingleMap(0.025, 1024, 1024, [2]float64{0, 0})
	m := &Map{
		Meta: &MapMetaData{
			Name:        "Singlemap",
			Description: "This single layered map is pretty rad.",
			MapType:     "logodds",
		},
		MapRep: mapRep,
	}

	err := m.Save("testoutput\\singlemap")
	if err != nil {
		t.Error(err)
	}
}

func TestLoadSingleMap(t *testing.T) {

	m, err := Load("testoutput\\singlemap")
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	t.Logf("Number of map levels in loaded maprep: %d", m.MapRep.GetMapLevels())

}

func TestLoadMakeFullImages(t *testing.T) {

	// Get list over maps
	mapList, err := GetMaps()
	if err != nil {
		t.Fatal(err)
	}

	for filename, _ := range mapList {
		fmt.Printf("Saving full image for %s ...", filename)

		// Load map
		m, err := Load(filename)
		if err != nil {
			t.Error(err)
		}

		fmt.Println("Map loaded.")

		// Free count
		freeCount := 0
		size := m.MapRep.GetGridMap(0).GetSizeX() * m.MapRep.GetGridMap(0).GetSizeY()
		for i := 0; i < size; i++ {
			if m.MapRep.GetGridMap(0).GetCellByIndex(i).IsFree() {
				freeCount++
			}
		}
		fmt.Printf("%d free cells.\n", freeCount)
		fmt.Println(m.MapRep.GetGridMap(0).GetMapDimProperties())

		// Get image
		img, err := mapimages.GetMapImage(m.MapRep)
		if err != nil {
			t.Error(err)
		}

		// Save the image
		file, err := os.Create("testoutput\\" + filename + ".png")
		defer file.Close()
		if err != nil {
			t.Error(err)
		}
		png.Encode(file, img)
		fmt.Println("Image saved.")
	}

}

func TestLoadMeta(t *testing.T) {

	meta, err := LoadMapMetaData("testoutput\\singlemap")
	if err != nil {
		t.Error(err)
	}

	if meta == nil {
		t.Errorf("Meta is nil")
	}

	t.Logf("Loaded map meta file: %s", meta.Name)

}

func TestLoadThumbnail(t *testing.T) {

	thumb, err := LoadMapThumbnail("testoutput\\singlemap")
	if err != nil {
		t.Error(err)
	}

	if thumb == nil {
		t.Errorf("Thumb is nil")
	}

	t.Logf("Loaded thumbnail with size: %d, %d", thumb.Bounds().Dx(), thumb.Bounds().Dy())

}

func TestEncodeGridMap(t *testing.T) {

	m := &Map{
		Meta:   &MapMetaData{},
		MapRep: maprep.MakeMapRepSingleMap(0.025, 1024, 1024, [2]float64{0, 0}),
	}

	gridmap := m.MapRep.GetGridMap(0)

	// Mark a cell so we can see that it's reloaded
	cell := gridmap.GetCell(0, 0).(*logoddsmap.LogOddsCell)
	cell.Set(1337)

	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(gridmap)
	if err != nil {
		t.Fatal(err)
	}

	// Read it back
	buf = bytes.NewBuffer(buf.Bytes())
	loaded := new(logoddsmap.OccGridMapLogOdds)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(loaded)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Loaded scaleToMap: %f", loaded.GetScaleToMap())
	t.Logf("Loaded sizes: %d, %d", loaded.GetSizeX(), loaded.GetSizeY())
	t.Logf("WorldTMap matrix: %s", loaded.GetWorldTmap())
	t.Logf("Cell (0, 0): %f", loaded.GetCell(0, 0).GetValue())
}

func TestGetMaps(t *testing.T) {

	maps, err := GetMaps()
	if err != nil {
		t.Error(err)
	}

	t.Logf("Found %d maps!", len(maps))

	for filename, m := range maps {
		t.Logf("%s: %s; %s", filename, m.Name, m.Description)
	}

}
