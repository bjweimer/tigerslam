// Package peasing implements a MotorDriver which controls the motors smoothly,
// i.e. by not setting the motors' speeds to the commanded speed right away,
// but with small intermediate updates. This makes the motors behave less
// quirky and may result in less wear and tear on mechanical parts.
//
// Easing is implemented by a P-controller. It also features a auto-stop
// feature, which
package peasing

import (
	"time"

	"robot/motor/driver"
	// "robot/motor/dummydriver"
)

// The interval at which the speeds of the motors are set
const UPDATE_INTERVAL = 100 * time.Millisecond

// If no new command comes within the time limit, the speed is set to 0
// (auto stop)
const AUTO_STOP_TIME = 1 * time.Second

// Turn auto-stop on/off
const USE_AUTO_STOP = true

// The P-value in the P-controller
const P_PARAMETER = 0.3

// Implements MotorDriver
type PEasingMotorDriver struct {
	driver          *driver.Motor
	stopChan        chan bool
	leftRef         float64
	rightRef        float64
	leftLast        float64
	rightLast       float64
	lastCommandTime time.Time
	err             error
}

func MakeDefaultPEasingMotorDriver() *PEasingMotorDriver {
	return &PEasingMotorDriver{
		driver: driver.MakeDefaultMotor(),
		// driver: &dummydriver.DummyDriver{},
	}
}

// Connect the motor, set up speed updating
func (q *PEasingMotorDriver) Connect() error {

	// Connect the physical motor
	err := q.driver.Connect()
	if err != nil {
		return err
	}

	// Set up the stop channel
	q.stopChan = make(chan bool)
	go q.updater()

	return nil
}

func (q *PEasingMotorDriver) Disconnect() error {
	err := q.driver.Disconnect()

	q.stopChan <- true

	return err
}

func (q *PEasingMotorDriver) IsConnected() bool {
	return q.driver.IsConnected()
}

// Set the reference speed for the left and right motors. The actual updates
// won't happen before the next cycle in the updating loop. This function
func (q *PEasingMotorDriver) SetSpeeds(left, right float64) error {

	if q.stopChan == nil {
		q.Connect()
	}

	// Set reference speed for the controller
	q.leftRef = left
	q.rightRef = right

	// Set time of last command (for auto-stop feature)
	q.lastCommandTime = time.Now()

	// If there has been an error since last time, return it even if it doesn't
	// have anything to do with *this* SetSpeeds() execution.
	if q.err != nil {
		t := q.err
		q.err = nil
		return t
	}

	return nil
}

// Continuously
func (q *PEasingMotorDriver) updater() {

	// Start a new ticker
	updateTicker := time.NewTicker(UPDATE_INTERVAL)

	for {

		// Check for stop command on channel
		select {
		case <-q.stopChan:
			q.stopChan = nil
			return
		default:
		}

		// Check if time's up for the command -- auto stop
		if USE_AUTO_STOP {
			if (q.leftRef != 0.0) && (q.rightRef != 0.0) && (time.Since(q.lastCommandTime) > AUTO_STOP_TIME) {
				q.autoStop()
			}
		}

		// Do our work. If the update fails, set error and continue.
		err := q.motorSpeedUpdate()
		if err != nil {
			q.err = err
		}

		// Wait for tick
		<-updateTicker.C

	}

}

// PID-update -- Implemented as a P-controller only (for now?)
//
// 		z_{k+1} = p * (r - z_k)
//
func (q *PEasingMotorDriver) motorSpeedUpdate() error {

	// Calculate the current speed
	left := q.leftLast + P_PARAMETER*(q.leftRef-q.leftLast)
	right := q.rightLast + P_PARAMETER*(q.rightRef-q.rightLast)

	// Push new speeds to motor
	err := q.driver.SetSpeeds(left, right)
	if err != nil {
		return err
	}

	// If we were successful, store values
	q.leftLast = left
	q.rightLast = right

	return nil
}

// Auto-stop (happens if there's no new command for AUTO_STOP_TIME) time.
func (q *PEasingMotorDriver) autoStop() {
	q.leftRef, q.rightRef = 0.0, 0.0
}
