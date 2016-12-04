package web

import (
	"net/http"
	"strings"
	"io/ioutil"
	
	auth "github.com/abbot/go-http-auth"
	
	"robot/config"
)

func (ws *WebServer) downloadHandler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	itemType := ws.getDownloadAction(r.URL.Path)
	data := r.URL.Query()
	
	switch itemType {
	case "log":
		file, err := ioutil.ReadFile(config.SENSORLOGS_ROOT + data.Get("file"))
		if err != nil {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}
		w.Header().Add("Content-Type", "text/csv")
		w.Header().Add("Content-Disposition", "attachment; filename=\"" + data.Get("file") + "\"")
		w.Write(file)
	default:
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

func (ws *WebServer) getDownloadAction(url string) string {
	return strings.TrimRight(url[len(ws.downloadURL):], "/")
}