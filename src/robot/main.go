package main

import (
	_ "expvar"
	"fmt"
	"log"
	"os"
	"runtime"

	"robot/controller"
	"robot/logging"
	"robot/logging/datalog"
	"robot/web"
)

var vmain bool = false

var logger *log.Logger

func main() {
	if vmain {
		fmt.Println("Entering Main")
	}
	// Set max CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	sigc := make(chan os.Signal, 1)
	logger = logging.New()
	logging.AddWriter(os.Stderr)

	// Set up state
	controller := controller.MakeDefaultController()

	// Set up web server
	ws := web.MakeWebServer(controller)
	logging.AddWriter(ws.GetLogWriter())
	datalog.AddWriter(ws.GetDataLogWriter())
	go ws.Serve()

	logger.Println("Ready to do awesome things.")

	// Wait for termination signal
	select {
	case <-sigc:
		logger.Println("Exiting.")
	}

	return
}
