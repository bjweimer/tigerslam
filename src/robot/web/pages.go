package web

import (
	"html/template"
	"net/http"
	"path/filepath"
	//	"strconv"
	"io/ioutil"
	"os"
	"path"
	"time"

	auth "github.com/abbot/go-http-auth"

	"robot/config"
	"robot/controller"
	"robot/mapstorage"
	"robot/slam"
)

// Describes a subpage (view)
type Page struct {
	templateName string
	funcs        template.FuncMap
	data         interface{}
}

// Common data for all pages
type Data struct {
	PAGE_NAME  string
	STATIC_URL string
	DATA       interface{}
}

func (ws *WebServer) pageHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	var page *Page
	var err error
	baseTemplatePath := ws.findTemplate(config.BASE_TEMPLATE)

	switch r.URL.Path {
	case "/manual-control/":
		page, err = manualControlPage()
	case "/sensors/":
		page, err = sensorsPage(ws)
	case "/slam/":
		page, err = slamPage(ws)
	case "/settings/":
		page, err = settingsPage()
	default:
		page, err = mainPage(ws)
	}
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Standard data and embed page data
	data := Data{
		page.templateName,
		ws.staticURL,
		page.data,
	}

	//
	funcMap := template.FuncMap{
		"equal":         equal,
		"smalldatetime": smalldatetime,
	}

	t := template.New(config.BASE_TEMPLATE)
	t.Funcs(funcMap).Funcs(page.funcs)
	_, err = t.ParseFiles(baseTemplatePath, ws.findTemplate(page.templateName))
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)

	if err != nil {
		logger.Println(err)
	}
}

func (ws *WebServer) findTemplate(template string) string {
	return filepath.Join(ws.templateRoot, template)
}

// Template equal function
func equal(a, b string) bool {
	return a == b
}

// Template byte to kilobyte function
func kilobyte(bytes int64) int64 {
	return bytes / 1024
}

// Small datetime
func smalldatetime(t time.Time) string {
	return t.Format(time.ANSIC)
}

// Main page
func mainPage(ws *WebServer) (*Page, error) {
	return &Page{
		"main.html",
		template.FuncMap{},
		struct {
			LOG_UPDATE_RATE int
			CONTROLLER      *controller.Controller
		}{
			config.LOG_UPDATE_RATE,
			ws.controller,
		},
	}, nil
}

// Manual control page
func manualControlPage() (*Page, error) {
	return &Page{
		"manual-control.html",
		template.FuncMap{},
		struct {
			MAX_SPEED float64
		}{
			config.MAX_SPEED,
		},
	}, nil
}

// Sensors page
func sensorsPage(ws *WebServer) (*Page, error) {

	// Get all log names
	files, err := ioutil.ReadDir(config.SENSORLOGS_ROOT)
	if err != nil {
		return nil, err
	}
	logs := make([]os.FileInfo, 0, len(files))
	for i := range files {
		if !files[i].IsDir() && path.Ext(files[i].Name()) == ".log" {
			logs = append(logs, files[i])
		}
	}

	return &Page{
		"sensors.html",
		template.FuncMap{
			"kilobyte": kilobyte,
			"equal":    equal,
		},
		struct {
			LOGS       []os.FileInfo
			CONTROLLER *controller.Controller
		}{
			logs,
			ws.controller,
		},
	}, nil
}

// SLAM page
func slamPage(ws *WebServer) (*Page, error) {
	var html string

	if ws.controller.SlamController.GetState() == slam.OFF {
		html = "slam-off.html"
	} else {
		html = "slam.html"
	}

	maps, err := mapstorage.GetMaps()
	if err != nil {
		maps = map[string]*mapstorage.MapMetaData{}
	}

	return &Page{
		html,
		template.FuncMap{},
		struct {
			LOG_UPDATE_RATE int
			CONTROLLER      *controller.Controller
			MAP_STORAGE     map[string]*mapstorage.MapMetaData
		}{
			config.LOG_UPDATE_RATE,
			ws.controller,
			maps,
		},
	}, nil
}

// Settings page
func settingsPage() (*Page, error) {

	// Make settings map from config, formatted like:
	// map[string]map{string}string
	// I.e. all values are strings

	settings := map[string]map[string]string{}
	sections := config.ConfigFile.GetSections()

	for _, section := range sections {
		settings[section] = map[string]string{}

		options, _ := config.ConfigFile.GetOptions(section)
		for _, option := range options {
			settings[section][option], _ = config.ConfigFile.GetString(section, option)
		}
	}

	// fmt.Println(settings["default"])

	return &Page{
		"settings.html",
		template.FuncMap{},
		settings["default"], // was just: settings,
	}, nil
}
