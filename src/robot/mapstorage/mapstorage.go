// Package mapstorage takes care of saving and loading maps.
//
// Maps are saved in special ZIP-compressed files, and consist of:
//  - Grid map data
//  - Meta data
//  - Map thumbnail
//
// These are saved together and can later be loaded.
package mapstorage

import (
	"archive/zip"
	"encoding/gob"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path"

	"hectormapping/map/mapimages"
	"hectormapping/map/maprep"

	"robot/config"
)

const MAP_FILE_EXTENSION = ".tigermap"
const MAPREP_SUBFILE_NAME = "map"
const MAPDATA_SUBFILE_NAME = "meta"
const THUMBNAIL_SUBFILE_NAME = "thumb.png"

// A Map is a grid map in the expanded sense, that is, complete with meta data.
type Map struct {
	Meta   *MapMetaData
	MapRep maprep.MapRepresentation
}

// Return a map with filenames as keys and meta data as descriptions.
func GetMaps() (map[string]*MapMetaData, error) {

	// Create the map
	maps := map[string]*MapMetaData{}

	// We must now look in config.MAP_STORAGE_ROOT for files with the correct
	// extension, then try to open them one by one and extract the meta data
	// information.
	files, err := ioutil.ReadDir(config.MAP_STORAGE_ROOT)
	if err != nil {
		return maps, err
	}

	for _, f := range files {
		if !f.IsDir() && path.Ext(f.Name()) == MAP_FILE_EXTENSION {

			// Well, this file isn't a directory and has the correct extension,
			// so it looks promising. Try to load it's meta data.
			meta, err := LoadMapMetaData(f.Name())
			if err != nil || meta == nil {
				break
			}

			// Oh joy, we found a map with proper meta data. Add it to the map!
			maps[f.Name()] = meta

		}
	}

	return maps, nil

}

// Get proper file path
// If no folder is specified, assume it's in the default map storage root.
// If no extension is specified, add the default extension.
func getFilePath(filename string) (filepath string) {

	// Find out if the filename specifies some folder. If not, save it in the
	// default folder (from Config).
	if path.Dir(filename) == "." {
		filepath = config.MAP_STORAGE_ROOT + filename
	} else {
		filepath = filename
	}

	// Find out if the filename specifies some file extension. If not, give it
	// the MAP_FILE_EXTENSION.
	if path.Ext(filepath) == "" {
		filepath += MAP_FILE_EXTENSION
	}

	return
}

// Save the entire Map object to a file
func (m *Map) Save(filename string) error {
	// Debug: fmt.Println("Entering Save map routine")
	// Create the file
	file, err := os.Create(getFilePath(filename))
	if err != nil {
		return err
	}
	// Debug: fmt.Println("File name:")
	// Debug: fmt.Println(file)

	// Create a new zip archive.
	archive := zip.NewWriter(file)

	// Set the flag telling if maprep is single - which it isn't (I think that's only for TinySLAM)??????????
	if _, ok := m.MapRep.(*maprep.MapRepSingleMap); ok {
		m.Meta.IsMapRepSingleMap = true
	}
	// Debug: fmt.Println("Is map single?:")
	// Debug: fmt.Println(m.Meta.IsMapRepSingleMap) -> false

	// Extract, save the map dimensions
	m.Meta.Mdp = m.MapRep.GetGridMap(0).GetMapDimProperties()
	// Debug: fmt.Printf("\nLet's look at this: m.Meta.Mdp = %+v\n\n", m.Meta.Mdp)

	// ******************seems like the map should be instantiated here to be saved?????????????????????????????????????????
	// Save the mapRep to archive
	m.saveMapDataToArchive(archive)
	m.saveMapRepToArchive(archive)
	m.saveThumbnailToArchive(archive)
	// Debug: fmt.Println("Saved entire map") //but an error prevents the map data being saved *********************

	// Close the archive
	err = archive.Close()
	if err != nil {
		return err
	}

	// Close the file
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

// Create the grid map file in the archive and write the map to it.
func (m *Map) saveMapRepToArchive(archive *zip.Writer) error {
	// Create the map file in the archive
	mapfile, err := archive.Create(MAPREP_SUBFILE_NAME)
	if err != nil {
		return err
	}

	// Create a gob encoder
	encoder := gob.NewEncoder(mapfile)

	// Debug: fmt.Println("We got this far - try to print the m.MapRep")
	//Debug: fmt.Printf("m.MapRep = %+v\n\n", m.MapRep)

	//goon.Dump(m.MapRep)

	// Encode the gridmap
	err = encoder.Encode(m.MapRep) // This is were the error occurs.
	if err != nil {
		fmt.Printf("An error occured in saveMapRepToArchive:\n%q\n", err) // This is the error!
		//fmt.Printf("This is the m.MapRep: \n %+v\n\n", m.MapRep)
		//fmt.Println("")
		//goon.Dump(m.MapRep.mapContainer)
		return err // This returns an error, so the grid map doesn't get saved!
	}

	return nil
}

// Load a complete Map object from file
func Load(filename string) (*Map, error) {

	m := &Map{}

	// Open the archive file
	r, err := zip.OpenReader(getFilePath(filename))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// We must first find the meta file
	for _, f := range r.File {
		if f.Name == MAPDATA_SUBFILE_NAME {
			m.loadMapDataFromFile(f)
		}
	}

	if m.Meta.Mdp == nil {
		return nil, errors.New("Could not get map dimensions.")
	}

	for _, f := range r.File {
		switch f.Name {
		case MAPREP_SUBFILE_NAME:
			err := m.loadMapRepFromFile(f, m.Meta.IsMapRepSingleMap)
			if err != nil {
				return nil, err
			}
		}
	}

	return m, nil

}

// Copy an archive to a new one
func Copy(filename, newName string) error {

	// Check if we shouldn't do anything
	if getFilePath(filename) == getFilePath(newName) {
		return nil
	}

	// Find out if a map with name newName already exists
	if _, err := os.Stat(getFilePath(newName)); err == nil {
		return errors.New("Map " + newName + " already exists.")
	}

	// Since the map file is an archive (zipped), we cannot simply change the
	// files in it. We copy everything over to a new one, and we want to do
	// this without actually parsing what's inside (loading map etc).
	newFile, err := os.Create(getFilePath(newName))
	if err != nil {
		return err
	}
	defer newFile.Close()

	// Open an archive writer for the new file
	newArchive := zip.NewWriter(newFile)

	// Open old archive
	oldArchive, err := zip.OpenReader(getFilePath(filename))
	if err != nil {
		return err
	}
	defer oldArchive.Close()

	// Transfer map and thumbnail
	moveSubFile(oldArchive, newArchive, MAPREP_SUBFILE_NAME)
	moveSubFile(oldArchive, newArchive, THUMBNAIL_SUBFILE_NAME)

	// Load curernt MapMetaData
	mmd, err := LoadMapMetaData(filename)
	if err != nil {
		return err
	}

	// Change name in mmd
	mmd.Name = newName

	// Create dummy map object with new meta data
	m := &Map{
		Meta: mmd,
	}

	// Save the meta data
	err = m.saveMapDataToArchive(newArchive)
	if err != nil {
		return err
	}

	err = newArchive.Close()
	if err != nil {
		// Should perhaps clean up by deleting the new archive file
		return err
	}

	return nil
}

// Delete a map file
func Delete(filename string) error {
	return os.Remove(getFilePath(filename))
}

// Rename a map file
func Rename(filename, newName string) error {
	err := Copy(filename, newName)
	if err != nil {
		return err
	}

	// Delete the old one and return
	return Delete(filename)
}

// Move a subfile from one archive to another
func moveSubFile(fromArchive *zip.ReadCloser, toArchive *zip.Writer, subFileName string) error {

	var fromFile *zip.File

	// Iterate through fromArchive until we find the file we want
	for _, f := range fromArchive.File {
		if f.FileInfo().Name() == subFileName {
			fromFile = f
		}
	}
	if fromFile == nil {
		return errors.New("Subfile " + subFileName + " not found.")
	}

	fromFileReader, err := fromFile.Open()
	if err != nil {
		return err
	}

	// Read old file into byte slice
	b, err := ioutil.ReadAll(fromFileReader)
	if err != nil {
		return err
	}

	// Create corresponding file in new archive
	toFile, err := toArchive.Create(subFileName)
	if err != nil {
		return err
	}

	// Write to the new file
	_, err = toFile.Write(b)
	if err != nil {
		return err
	}

	return nil

}

// Load MapMetaData object only from archive file
func LoadMapMetaData(filename string) (*MapMetaData, error) {

	m := &Map{}

	// Open the archive file
	r, err := zip.OpenReader(getFilePath(filename))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Locate, load meta data file
	for _, f := range r.File {
		if f.Name == MAPDATA_SUBFILE_NAME {
			m.loadMapDataFromFile(f)
		}
	}

	// Go ahead and return it
	return m.Meta, nil
}

// Load map thumbnail image only from archive file
func LoadMapThumbnail(filename string) (image.Image, error) {

	var thumb image.Image
	var err error

	// Open the archive file
	r, err := zip.OpenReader(getFilePath(filename))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Locate, load thumbnail file
	for _, f := range r.File {
		if f.Name == THUMBNAIL_SUBFILE_NAME {

			// Open file
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}

			thumb, err = png.Decode(rc)
			if err != nil {
				return nil, err
			}

			rc.Close()
		}
	}

	if thumb == nil {
		return nil, errors.New("Could not load thumbnail")
	}

	return thumb, nil
}

// Create the map data file in the archive and write data to it.
func (m *Map) saveMapDataToArchive(archive *zip.Writer) error {

	// Create the map data file in the archive
	datafile, err := archive.Create(MAPDATA_SUBFILE_NAME)
	if err != nil {
		return err
	}

	// Create a gob encoder
	encoder := gob.NewEncoder(datafile)

	// Encode the MapMetaData
	err = encoder.Encode(m.Meta)
	if err != nil {
		fmt.Printf("An error occured in saveMapDataToArchive:\n%v\n", err)
		return err
	}

	return nil
}

// Create a thumbnail PNG of the mapRep and save it to the archive.
func (m *Map) saveThumbnailToArchive(archive *zip.Writer) error {

	// Create thumbnail file in the archive
	thumbfile, err := archive.Create(THUMBNAIL_SUBFILE_NAME)
	if err != nil {
		return err
	}

	// Get thumbnail: zoomLevel 0, coords (0, 0)
	thumb, err := mapimages.GetMapTile(m.MapRep, 0, 0, 0)
	if err != nil {
		return err
	}

	// Encode the image to file
	err = png.Encode(thumbfile, thumb)
	if err != nil {
		fmt.Printf("An error occured in saveThunbnailToArchive:\n%v\n", err)
		return err
	}

	return nil
}

// Load the map data file in the archive
func (m *Map) loadMapDataFromFile(file *zip.File) error {

	// Open the file in the archive
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create gob decoder
	decoder := gob.NewDecoder(rc)

	err = decoder.Decode(&m.Meta)
	if err != nil {
		return err
	}

	return nil
}

// Load the grid map file in the archive
func (m *Map) loadMapRepFromFile(file *zip.File, singleMap bool) error {
	//Debug: fmt.Println("Enter loadMapRepFromFile")
	// Open the file in the archive
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create a gob decoder
	decoder := gob.NewDecoder(rc)

	if singleMap {
		//Debug: fmt.Println("singleMap")
		// Decode the singlemap
		var mapRep maprep.MapRepSingleMap
		err = decoder.Decode(&mapRep)
		if err != nil && err != io.EOF {
			return err
		}
		m.MapRep = &mapRep

	} else {
		//Debug: fmt.Println("multiMap")
		// Decode the multimap
		var mapRep maprep.MapRepMultiMap
		err = decoder.Decode(&mapRep)
		if err != nil && err != io.EOF {
			return err
		}
		m.MapRep = &mapRep
		//Debug: fmt.Println(&mapRep) loading Gary's map returns two empty slices

	}

	return nil
}
