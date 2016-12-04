package datalog

type Writer interface {
	Write(key string, data ... float64)
}

type multiplexedDataWriter struct {
	writers []Writer
}

func (mw *multiplexedDataWriter) Write(key string, data ... float64) {
	for _, w := range(mw.writers) {
		w.Write(key, data...)
	}
}

var writer *multiplexedDataWriter

type Logger struct {
	writer Writer
}

func (d *Logger) Log(key string, data ... float64) {
	writer.Write(key, data...)
}

func New() *Logger {
	return &Logger{writer:writer}
}

func AddWriter(w Writer) {
	writer.writers = append(writer.writers, w)
}

func init() {
	// Create the logger
	writer = new(multiplexedDataWriter)
}

