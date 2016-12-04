package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"robot/config"
	"robot/controller"
	"robot/mapstorage"
	// "robot/slam"

	auth "github.com/abbot/go-http-auth"
)

var API_FUNCMAP = map[string]func(http.ResponseWriter, *controller.Controller, url.Values) ([]byte, error){
	// "get/log": getLog,

	"set/slam/initialize":                 setSlamInitialize,
	"set/slam/initialize-from-stored-map": setSlamInitializeFromStoredMap,
	"set/slam/start":                      setSlamStart,
	"set/slam/stop":                       setSlamStop,
	"set/slam/terminate":                  setSlamTerminate,
	"set/slam/save":                       setSlamSave,
	"get/slam/image/full":                 getSlamImageFull,
	"get/slam/image/tile":                 getSlamImageTile,
	"get/slam/stats":                      getSlamStats,

	"get/mapstorage/package":   getMapstoragePackage,
	"get/mapstorage/metadata":  getMapstorageMetadata,
	"get/mapstorage/thumbnail": getMapstorageThumbnail,
	"set/mapstorage/mapname":   setMapstorageMapname,

	"set/sensors/connect":    SetSensorsConnect,
	"set/sensors/disconnect": SetSensorsDisconnect,
	"set/sensors/start":      SetSensorsStart,
	"set/sensors/stop":       SetSensorsStop,

	"set/sensorlogs/delete":                 SetSensorlogsDelete,
	"set/sensorlogs/rename":                 SetSensorlogsRename,
	"set/sensorlogs/start-logread-realtime": SetSensorlogsStartlogreadrealtime,
	"set/sensorlogs/stop-logread-realtime":  SetSensorlogsStoplogreadrealtime,

	"set/motor/disconnect":        SetMotorDisconnect,
	"set/motor/speeds":            SetMotorSpeeds,
	"set/motor/planpath":          SetMotorPlanPath,
	"get/motor/path":              GetMotorPath,
	"set/motor/followpath":        SetMotorFollowPath,
	"set/motor/stoppathfollowing": SetMotorStopPathFollowing,
	"set/motor/deletepath":        SetMotorDeletePath,
}

func (ws *WebServer) getAPIAction(url string) string {
	return strings.TrimRight(url[len(ws.apiURL):], "/")
}

// Handle API requests
//
// Execute the correct API function.
func (ws *WebServer) apiHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {

	action := ws.getAPIAction(r.URL.Path)
	data := r.URL.Query()

	// This is dirty and should be fixed: log needs ws' log
	if action == "get/log" {
		ws.getLog(w, ws.controller, data)
		return
	}

	apiFunc, ok := API_FUNCMAP[action]
	if !ok {
		http.Error(w, "Unrecognized API action:"+action, http.StatusBadRequest)
		return
	}

	resp, err := apiFunc(w, ws.controller, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusExpectationFailed)
		return
	}

	w.Write(resp)
}

func (ws *WebServer) getLog(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) {
	resp, err := json.Marshal(ws.log)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}

// Initialize a slam algorithm.
func setSlamInitialize(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {
	alg := data.Get("algorithm")

	err := ctrl.SlamController.InitializeSlam(alg, ctrl.Robot)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

// This function will block for quite some time.
func setSlamInitializeFromStoredMap(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {
	alg := data.Get("algorithm")
	filename := data.Get("filename")
	//Debug: fmt.Println(filename)
	err := ctrl.SlamController.InitializeSlamFromStoredMap(filename, alg, ctrl.Robot)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	//Debug: fmt.Println("setSlamInitialize completed") it works to here
	return json.Marshal("ok")
}

// Start an initialized SLAM algorithm
func setSlamStart(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {
	err := ctrl.SlamController.StartSlam()
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

// Stop a running SLAM algorithm
func setSlamStop(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	err := ctrl.SlamController.StopSlam()
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

// Terminate a SLAM algorithm (i.e. release release its memory, un-initialize it)
func setSlamTerminate(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	ctrl.SlamController.TerminateSlam()

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

// Save the map of the currently initialized SLAM algorithm to a map
// package.
func setSlamSave(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	// Get the MapRepresentation from the SlamController module, then save it
	// with meta data.
	fmt.Println("Enter setSlamSave function")

	mapName := data.Get("name")
	mapDescription := data.Get("description")

	if mapName == "" {
		return nil, errors.New("No map name specified")
	}

	m := &mapstorage.Map{
		Meta: &mapstorage.MapMetaData{
			Name:        mapName,
			Description: mapDescription,
		},
		MapRep: ctrl.SlamController.GetMapRepresentation(),
	}
	//Debug: fmt.Printf("url.Values = %+v\n", data)
	//Debug: fmt.Println("mapName = ", mapName)
	//Debug: fmt.Println("mapDescription = ", mapDescription)
	//Debug: fmt.Printf("MapRep = %+v\n\n", ctrl.SlamController.GetMapRepresentation())

	filename := mapName + mapstorage.MAP_FILE_EXTENSION

	err := m.Save(filename)
	if err != nil {
		fmt.Printf("Error while saving:\n%v\n\n", err)
		return nil, errors.New("Error while saving")
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func getSlamImageFull(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {
	//Debug: fmt.Println("getSlamImageFull")
	if ctrl.SlamController.GetSlam() == nil {
		return nil, errors.New("SLAM not initialized.")
	}

	img, err := ctrl.SlamController.GetSlam().GetMapImage()
	if err != nil {
		return nil, err
	}

	png.Encode(w, img)

	w.Header().Add("Content-Type", "image/png")

	return nil, nil
}

func getSlamImageTile(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {
	//Debug: fmt.Println("getSlamImageTile")
	if ctrl.SlamController.GetSlam() == nil {
		return nil, errors.New("SLAM not initialized.")
	}

	zoomLevel, err := strconv.Atoi(data.Get("zoomLevel"))
	if err != nil {
		return nil, err
	}
	tileX, err := strconv.Atoi(data.Get("tileX"))
	if err != nil {
		return nil, err
	}
	tileY, err := strconv.Atoi(data.Get("tileY"))
	if err != nil {
		return nil, err
	}

	img, err := ctrl.SlamController.GetSlam().GetMapTile(uint(zoomLevel), tileX, tileY)
	if err != nil {
		return nil, err
	}

	png.Encode(w, img)

	w.Header().Add("Content-Type", "image/png")

	return nil, nil
}

func getMapstoragePackage(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {
	//Debug: fmt.Println("getMapstoragePackage")
	filename := data.Get("filename")
	f, err := os.Open(config.MAP_STORAGE_ROOT + filename)
	if err != nil {
		return nil, errors.New("No such map package")
	}

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		if _, err = w.Write(buf[:n]); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func getMapstorageMetadata(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	filename := data.Get("filename")

	meta, err := mapstorage.LoadMapMetaData(filename)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal(meta)
}

func getMapstorageThumbnail(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	filename := data.Get("filename")

	thumbnail, err := mapstorage.LoadMapThumbnail(filename)
	if err != nil {
		return nil, err
	}

	png.Encode(w, thumbnail)
	w.Header().Add("Content-Type", "image/png")

	return nil, nil
}

func setMapstorageMapname(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	filename := data.Get("filename")
	newName := data.Get("newname")

	err := mapstorage.Rename(filename, newName)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

// Returns stats relevant for the SLAM page -- extracted from SLAM and motor controllers
func getSlamStats(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	if ctrl.SlamController.GetSlam() == nil {
		return nil, errors.New("SLAM not initialized.")
	}

	stats := map[string]interface{}{
		"state":      ctrl.SlamController.GetState(),
		"position":   ctrl.SlamController.GetSlam().GetPosition(),
		"motorState": ctrl.MotorController.GetStateString(),
	}

	path := ctrl.MotorController.GetPath()
	if path == nil {
		stats["motorPathID"] = nil
	} else {
		stats["motorPathID"] = path.ID
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal(stats)
}

func SetSensorsConnect(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	sensor := data.Get("sensor")

	err := ctrl.SensorController.ConnectSensor(sensor)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorsDisconnect(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	sensor := data.Get("sensor")

	err := ctrl.SensorController.DisconnectSensor(sensor)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorsStart(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	sensor := data.Get("sensor")

	err := ctrl.SensorController.StartSensor(sensor)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorsStop(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	sensor := data.Get("sensor")

	err := ctrl.SensorController.StopSensor(sensor)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorlogsDelete(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	logName := data.Get("logname")

	err := os.Remove(config.SENSORLOGS_ROOT + logName)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorlogsRename(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	logName := data.Get("logname")
	newName := data.Get("newname")

	err := os.Rename(config.SENSORLOGS_ROOT+logName, config.SENSORLOGS_ROOT+newName+".log")
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorlogsStartlogreadrealtime(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	logName := data.Get("logname")
	err := ctrl.SensorController.StartLogReadRealtime(logName)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetSensorlogsStoplogreadrealtime(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	ctrl.SensorController.StopLogReadRealtime()

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetMotorDisconnect(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	err := ctrl.MotorController.Disconnect()
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetMotorSpeeds(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	left, err := strconv.ParseFloat(data.Get("left"), 64)
	if err != nil {
		return nil, errors.New("Invalid data")
	}

	right, err := strconv.ParseFloat(data.Get("right"), 64)
	if err != nil {
		return nil, errors.New("Invalid data")
	}

	err = ctrl.MotorController.ManualSpeeds(left, right)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetMotorPlanPath(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	x, err := strconv.ParseFloat(data.Get("x"), 64)
	if err != nil {
		return nil, errors.New("Invalid data")
	}

	y, err := strconv.ParseFloat(data.Get("y"), 64)
	if err != nil {
		return nil, errors.New("Invalid data")
	}

	err = ctrl.PlanPath([3]float64{x, y, 0})
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func GetMotorPath(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	path := ctrl.MotorController.GetPath()

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal(path)
}

func SetMotorFollowPath(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	err := ctrl.FollowPath()
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")

}

func SetMotorStopPathFollowing(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	ctrl.MotorController.StopPathFollowing()

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}

func SetMotorDeletePath(w http.ResponseWriter, ctrl *controller.Controller, data url.Values) ([]byte, error) {

	ctrl.MotorController.DeletePath()

	w.Header().Add("Content-Type", "application/json")
	return json.Marshal("ok")
}
