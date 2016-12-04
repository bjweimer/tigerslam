package web

import (
	"io"
	"log"
	"net/http"
	"path/filepath"

	auth "github.com/abbot/go-http-auth"
	//	"code.google.com/p/go.net/websocket"
	"golang.org/x/net/websocket"

	"robot/config"
	"robot/controller"
	"robot/logging"
	"robot/logging/datalog"
)

var authenticator *auth.BasicAuth
var logger *log.Logger

type WebServer struct {
	apiURL       string
	staticURL    string
	downloadURL  string
	staticRoot   string
	templateRoot string
	log          LogBuffer
	datalog      *DataWriter
	controller   *controller.Controller
}

func Secret(user, realm string) string {
	if user == config.USERNAME {
		return config.PASSPHRASE
	}
	return ""
}

func MakeWebServer(controller *controller.Controller) (ws *WebServer) {
	logger = logging.New()

	ws = &WebServer{
		apiURL:       config.API_URL,
		staticURL:    config.STATIC_URL,
		downloadURL:  "/download/",
		staticRoot:   filepath.Dir(config.STATIC_ROOT),
		templateRoot: filepath.Dir(config.TEMPLATE_ROOT),
		log:          MakeLogBuffer(),
		datalog:      NewDataWriter(),
		controller:   controller,
	}

	authenticator = auth.NewBasicAuthenticator(config.REALM, Secret)

	return
}

func (ws *WebServer) GetLogWriter() io.Writer {
	return &ws.log
}

func (ws *WebServer) GetDataLogWriter() datalog.Writer {
	return ws.datalog
}

func (ws *WebServer) Serve() {
	http.HandleFunc(ws.staticURL, func(w http.ResponseWriter, r *http.Request) {
		ws.staticHandler(w, r)
	})
	http.Handle("/api/streaming/log/", websocket.Handler(func(w *websocket.Conn) { ws.logStreamingServer(w) }))
	http.HandleFunc(ws.apiURL, authenticator.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		ws.apiHandler(w, r)
	}))
	http.HandleFunc("/download/", authenticator.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		ws.downloadHandler(w, r)
	}))
	http.HandleFunc("/", authenticator.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		ws.pageHandler(w, r)
	}))

	logger.Printf("Web server now serving at %s!", config.WEB_ADDRESS)
	http.ListenAndServe(config.WEB_ADDRESS, nil)
}

// Handler for static files
func (ws *WebServer) staticHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(ws.staticRoot, r.URL.Path[len(ws.staticURL):])
	http.ServeFile(w, r, path)
}
