// Package driver does communication to the robot motor driver card.
//
// It communicates over a serial interface.
package driver

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	// Original: serial "github.com/tarm/goserial"
	"github.com/tarm/serial"

	"robot/config"
	"robot/logging"
	"robot/tools/intmath"
)

type Direction int
type Side int

var logger *log.Logger

const (
	RANGE = 2000
)

const (
	DIRECTION_FORWARD = iota
	DIRECTION_BACKWARD
)

const (
	SIDE_LEFT = iota
	SIDE_RIGHT
)

// The motor struct holds parameters for communicating with the motor driver
// card and the driver card's parameters, and implements methods for
// commuication.
type Motor struct {
	config      *serial.Config
	port        io.ReadWriteCloser
	speed_left  int
	speed_right int
}

func init() {
	logger = logging.New()
}

// Make an arbitrary motor.
func MakeMotor(config *serial.Config) *Motor {
	return &Motor{
		config: config,
	}
}

// Make default motor from standard settings and config file.
func MakeDefaultMotor() *Motor {
	return &Motor{
		config: &serial.Config{
			Name: config.MOTORS_COM_NAME,
			Baud: config.MOTORS_BAUD_RATE,
		},
	}
}

// Set up the connection to the motor driver card over serial interface.
func (m *Motor) Connect() error {
	// Original: s, err := serial.OpenPort(m.config)
	logger.Printf("Motor connected on COM = %v\n", config.MOTORS_COM_NAME)
	//c := &serial.Config{Name: "COM6", Baud: 115200}
	c := &serial.Config{Name: config.MOTORS_COM_NAME, Baud: 115200}
	s, err := serial.OpenPort(c)
	time.Sleep(time.Second)
	if err != nil {
		return err
	}

	// logger.Println("Motor connected.")
	// fmt.Printf("COM = %q\n", m.config.Name)
	// fmt.Printf("Baud = %d\n\n", m.config.Baud)
	m.port = s

	return nil
}

// Disconnect
func (m *Motor) Disconnect() error {
	if m.port == nil {
		return nil
	}

	err := m.port.Close()
	if err != nil {
		return err
	}

	logger.Println("Motor disconnected.")
	m.port = nil
	return nil
}

func (m *Motor) IsConnected() bool {
	return m.port != nil
}

// Set the speeds of the left and right motors, using floating point numbers
// between -1 and 1, indicating speeds within the range of the motors. 1
// indicates the maximum speed, while 0 is stopped and -1 is maximum backward
// speed.
func (m *Motor) SetSpeeds(leftDecimal, rightDecimal float64) error {

	if !m.IsConnected() {
		err := m.Connect()
		if err != nil {
			return err
		}
	}

	// Incoming are [-1, 1]{float}. Convert to [-RANGE, RANGE]{int}

	// Original line: leftInt := int(leftDecimal * float64(RANGE))
	// Original line: rightInt := int(rightDecimal * float64(RANGE))

	leftInt := int(((leftDecimal+1.0)/2.0)*126) + 1     // range: 1 - 127
	rightInt := int(((rightDecimal+1.0)/2.0)*126) + 128 // range: 128 - 255

	//Debug: fmt.Printf("Left = %d", leftInt)
	//Debug: fmt.Printf("Right = %d\n", rightInt)

	// Check if we're inside limits
	// if intmath.Abs(leftInt) > RANGE || intmath.Abs(rightInt) > RANGE {
	// 	return errors.New("Outside range")
	// }

	// Cut to within range
	// Original: leftInt = intmath.Max(intmath.Min(leftInt, RANGE), -RANGE)
	// Original: rightInt = intmath.Max(intmath.Min(rightInt, RANGE), -RANGE)

	leftInt = intmath.Max(intmath.Min(leftInt, 127), 1)
	rightInt = intmath.Max(intmath.Min(rightInt, 255), 128)

	b1 := make([]byte, 1)
	b4 := make([]byte, 4)

	binary.LittleEndian.PutUint32(b4, uint32(leftInt))

	b1[0] = b4[0]

	//fmt.Println("leftInt = ", b1)

	_, err := m.port.Write([]byte(b1))
	if err != nil {
		fmt.Println("Error sending left wheel value: ", err)
		return err
	}

	binary.LittleEndian.PutUint32(b4, uint32(rightInt))

	b1[0] = b4[0]

	//fmt.Println("rightInt = ", b1)

	_, err = m.port.Write([]byte(b1))
	if err != nil {
		fmt.Println("Error sending right wheel value: ", err)
		return err
	}

	/*
		var err error

		// Sets the left side
		err = m.setSideSpeed(SIDE_LEFT, leftInt)
		if err != nil {
			m.Disconnect()
			return err
		}

		// Sets the right side
		err = m.setSideSpeed(SIDE_RIGHT, rightInt)
		if err != nil {
			m.Disconnect()
			return err
		}
	*/
	return nil
}

/*
// Set speed of one motor with implicit direction (sign of speed)
func (m *Motor) setSideSpeed(side Side, speed int) error {

	var direction Direction
	if speed >= 0 {
		direction = DIRECTION_FORWARD
	} else {
		direction = DIRECTION_BACKWARD
	}

	return m.setSpeedWithDirection(side, direction, intmath.Abs(speed))
}

// Set the speed of one motor with direction as a parameter
func (m *Motor) setSpeedWithDirection(side Side, direction Direction, speed int) error {

	if !m.IsConnected() {
		return errors.New("Motor not connected.")
	}

	// Check if speed is allowable
	if intmath.Abs(speed) > RANGE {
		return errors.New("Speed outside allowable range")
	}

	var num int
	if direction == DIRECTION_FORWARD {
		num = speed
	} else {
		num = -speed
	}

	speed = speed / 2

	// Communicate with motor card
	com := fmt.Sprintf("%s%s.%s", side, direction, speedFormat(speed))
	err := m.write(com)
	if err != nil {
		return err
	}

	// Write the change to motor
	if side == SIDE_LEFT {
		m.speed_left = num
	} else {
		m.speed_right = num
	}

	return nil
}

// Write string to the serial connection
func (m *Motor) write(s string) error {
	//fmt.Printf("s = %s", s)
	//fmt.Println()
	_, err := m.port.Write([]byte(s))
	return err
}

// Read string from the serial connection
func (m *Motor) read() (string, error) {
	buf := make([]byte, 128)
	n, err := m.port.Read(buf)
	return string(buf[:n]), err
}

func (s Side) String() string {
	if s == SIDE_LEFT {
		return "v"
	}
	return "h"
}

func (d Direction) String() string {
	if d == DIRECTION_FORWARD {
		return "+"
	}
	return "-"
}

func speedFormat(speed int) string {
	return fmt.Sprintf("%.4d", speed)
}
*/
