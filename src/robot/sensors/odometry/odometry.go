package odometry

import (
	"errors"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	serial "github.com/tarm/goserial"

	"robot/config"
	"robot/fsm"
	"robot/logging"
	"robot/model"
	"robot/sensors/sensor"
)

const NAME = "ODOMETRY"

var OdometrySensor *Encoder

// Pattern for reading from encoder card
const patternStr = `H:(-?\d+) V:(-?\d+);`

var pattern *regexp.Regexp

var logger *log.Logger

func init() {
	OdometrySensor = MakeDefaultEncoder()

	pattern = regexp.MustCompile(patternStr)

	logger = logging.New()
}

type Encoder struct {
	sensor.BasicSensor
	config   *serial.Config
	port     io.ReadWriteCloser
	stopChan chan bool
	ticker   *time.Ticker
	robot    *model.DifferentialWheeledRobot
}

// Make an arbitrary encoder
func MakeEncoder(config *serial.Config) *Encoder {
	e := &Encoder{
		BasicSensor: sensor.BasicSensor{
			FSM:  *fsm.MakeFSM(sensor.OFF),
			Name: NAME,
		},
		config: config,
		robot:  model.MakeDefaultDifferentialWheeledRobot(),
	}

	return e
}

// Make encoder from default parameters and config file
func MakeDefaultEncoder() *Encoder {
	conf := &serial.Config{
		Name: config.ODOMETRY_COM_NAME,
		Baud: config.ODOMETRY_BAUD_RATE,
	}
	return MakeEncoder(conf)
}

// Return all parameters as a string-interface map
func (e Encoder) GetParameters() map[string]interface{} {
	return map[string]interface{}{}
}

func (e *Encoder) Connect() error {
	s, err := serial.OpenPort(e.config)
	if err != nil {
		return err
	}

	logger.Println("Encoder connected.")
	e.SetState(sensor.CONNECTED)
	e.port = s
	return nil
}

func (e *Encoder) Disconnect() {
	if e.port == nil {
		return
	}

	e.port.Close()

	logger.Println("Encoder disconnected.")
	e.SetState(sensor.OFF)
	e.port = nil
	return
}

func (e *Encoder) Start() error {

	if e.GetState() != sensor.CONNECTED {
		return errors.New("Encoder not connected.")
	}

	e.ticker = time.NewTicker(100 * time.Millisecond)

	logger.Println("Encoder running.")
	e.SetState(sensor.RUNNING)

	e.stopChan = make(chan bool, 1)

	// Continuously read values and distribute them
	go func() {
		for {

			// Wait for tick or stop
			select {
			case <-e.stopChan:
				logger.Println("Encoder stopped.")
				e.SetState(sensor.CONNECTED)
				e.ticker.Stop()
				return
			case <-e.ticker.C:
				// noop
			}

			// Get measurement
			reading, err := e.GetData()
			if err != nil {
				logger.Println(err)
			} else {

				// Publish
				e.Distribute(reading)

			}

		}
	}()

	return nil
}

func (e *Encoder) Stop() {
	e.stopChan <- true
}

// Request odometric data from encoder card, with "l" (lowercase L) command,
// then get and interpret the answer.
// The answer is encoded as H:([-\d]+) V:([-\d]+);
func (e *Encoder) GetData() (*OdometryReading, error) {

	// Send command to encoder, telling it to return a reading
	err := e.write("l")
	if err != nil {
		return nil, err
	}

	// Retrieve a string, keep reading until we get a ";"
	var s string
	for !strings.Contains(s, ";") {
		sbuf, err := e.read()
		if err != nil {
			return nil, err
		}
		s += sbuf
	}

	leftPulses, rightPulses, err := e.parseAnswer(s)
	if err != nil {
		return nil, err
	}

	// Wrap in an OdometryReading
	or := MakeOdometryReading()
	or.LeftPulses = leftPulses
	or.RightPulses = rightPulses
	or.SetTimestamp(time.Now())

	return or, nil
}

// Write string to encoder card
func (e *Encoder) write(s string) error {
	_, err := e.port.Write([]byte(s))
	return err
}

// Read string from encoder card
func (e *Encoder) read() (string, error) {
	buf := make([]byte, 128)
	n, err := e.port.Read(buf)
	return string(buf[:n]), err
}

// Parse answer to pulse readings for left and right encoder.
func (e *Encoder) parseAnswer(answer string) (left int, right int, err error) {

	// Interpret the string
	matches := pattern.FindStringSubmatch(answer)
	if matches == nil {
		return 0, 0, errors.New("Invalid response from encoder: " + answer)
	}

	// Matches now contains two integers (might be negative) as strings; right
	// first.
	rightPulses, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	leftPulses, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return int(leftPulses), int(rightPulses), nil

}
