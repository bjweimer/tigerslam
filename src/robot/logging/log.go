package logging

import (
	"fmt"
	"log"
	"os"
	"io"
)

/*** Logger  ***/
var writer *multiplexedWriter

func init() {
	// Create the logger
	writer = new(multiplexedWriter)
}

func New() *log.Logger {
	return log.New(writer, "", log.Ltime|log.Lshortfile)
}

func AddWriter(w io.Writer) {
	writer.writers = append(writer.writers, w)
}

// Prints and writes the log stream to a file.
type multiplexedWriter struct {
	writers []io.Writer
}


func (mw *multiplexedWriter) Write(data []byte) (n int, err error) {
	var m, dm int
	n = len(data)
	
	for i, w := range(mw.writers) {
		m = 0
		for m < n {
			dm, err = w.Write(data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "log error: Write failes for writer[%d]: '%v'", i, w)
				break
			}
			m += dm
		}
	}
	return
}
