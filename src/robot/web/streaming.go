package web

import (
	//	"os"
	//	"time"
	"encoding/json"
	"robot/logging"
	//	"code.google.com/p/go.net/websocket"
	"golang.org/x/net/websocket"
)

// The StreamingLog type implements io.Writer
type StreamingLogWriter struct {
	out chan string
}

func MakeStreamingLogWriter(out chan string) *StreamingLogWriter {
	return &StreamingLogWriter{out}
}

func (slw *StreamingLogWriter) Write(p []byte) (n int, err error) {
	slw.out <- string(p)
	return len(p), nil
}

func (ws *WebServer) logStreamingServer(conn *websocket.Conn) {
	var message = map[string]string{
		"type": "logEntry",
	}
	var err error
	var json_txt []byte

	channel := make(chan string)
	streamingLogWriter := MakeStreamingLogWriter(channel)
	logging.AddWriter(streamingLogWriter)

	for {
		message["data"] = <-channel

		json_txt, err = json.Marshal(message)
		if err != nil {
			logger.Println(err.Error())
		}

		_, err = conn.Write(json_txt)
		if err != nil {
			conn.Close()
		}
	}
}
