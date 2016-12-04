// The Sensor package is responsible for implementing features common for every
// sensor and sensor reading. It standardizes subscribtions on channels and
// provides a common sensor interface for the rest of the program.
package sensor

import (
	"log"

	"robot/fsm"
	"robot/logging"
)

var logger *log.Logger

const (
	OFF       = "OFF"
	CONNECTED = "CONNECTED"
	RUNNING   = "RUNNING"
)

func init() {
	logger = logging.New()
}

type Sensor interface {
	Subscribe() chan SensorReading
	Unsubscribe(chan SensorReading)
	GetTypeName() string
	GetParameters() map[string]interface{}
	Connect() error
	Disconnect()
	Start() error
	Stop()
	GetState() fsm.State
	GetStateString() string
}

type BasicSensor struct {
	fsm.FSM
	Name        string
	subscribers []chan SensorReading
}

// Start a channel distributing readings
func (bs *BasicSensor) Subscribe() chan SensorReading {
	ch := make(chan SensorReading)
	bs.subscribers = append(bs.subscribers, ch)
	return ch
}

// Unsubscribe
func (bs *BasicSensor) Unsubscribe(ch chan SensorReading) {
	for i := range bs.subscribers {
		if ch == bs.subscribers[i] {
			// remove it
			bs.subscribers = append(bs.subscribers[:i], bs.subscribers[i+1:]...)
			return
		}
	}
}

// Close all channels
func (bs *BasicSensor) CloseSubscribtions() {
	for i := range bs.subscribers {
		close(bs.subscribers[i])
	}
}

// Distribute a reading among the subscribers
func (bs *BasicSensor) Distribute(r SensorReading) {
	for i := range bs.subscribers {
		// bs.subscribers[i] <- r
		select {
		case bs.subscribers[i] <- r:
			// Reading now sent
		default:
			// Take measures if sending was blocked
			//			<-bs.subscribers[i] // discard
			//			bs.subscribers[i] <- r // put in new
			logger.Printf("Sensor %s output channel blocked, reading discarded", bs.GetTypeName())
		}

	}
}

// Return the type name
func (bs *BasicSensor) GetTypeName() string {
	return bs.Name
}
