package web

import "sync"

type LogBuffer []string

func MakeLogBuffer() LogBuffer {
	return make([]string, 0, 255)
}

func (l *LogBuffer) Write(b []byte) (int, error) {
	*l = append(*l, string(b))
	return len(b), nil
}

// Used for ploting data
type DataWriter struct {
	sync.RWMutex
	titles []string
	data   [][]float64
}

func NewDataWriter() (d *DataWriter) {
	d = new(DataWriter)
	d.titles = make([]string, 0, 3)
	d.data   = make([][]float64, 0, 3)
	return
}

func (d *DataWriter) Write(key string, y ...float64) {
	d.Lock()
	defer d.Unlock()

	var index int = -1
	for i, title := range d.titles {
		if title == key {
			index = i
			break
		}
	}
	if index == -1 {
		d.titles = append(d.titles, key)
		d.data = append(d.data, y)
	} else {
		d.data[index] = y
	}
}
