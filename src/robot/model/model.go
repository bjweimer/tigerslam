// Package model implements a mathematical model of the robot
package model

import (
	"fmt"
	"math"

	"robot/config"
)

// Model implements different kinds of robots
type Robot interface {
	TypeString() string
	RollPosition(float64, float64, Position) Position
	OdometryPosition(int, int, Position) Position
	GetParameters() interface{}
}

// A position with orientation, in meters and radians.
type Position struct {
	X, Y  float64
	Theta float64
}

func (p Position) String() string {
	return fmt.Sprintf("X: %f, Y: %f, Theta: %f", p.X, p.Y, p.Theta)
}

// DifferentialWheeledRobot implements a model of a robot which has two
// parallel-mounted wheels. All sizes are in SI units. If both wheels rotate
// with the same speed, the robot goes straight ahead. If the speeds are
// different, it rotates about the central point of the wheel axis.
type DifferentialWheeledRobot struct {
	// The distance between the the centers of the two wheels.
	BaseWidth float64
	// The radius of the wheels
	WheelRadius float64
	// LeftWheelRadius = RightWheelRadius * Ratio (typically 1) -- NOT YET IMPLEMENTED
	WheelRatio float64
	// Pulses per rotation in odometry
	OdometryPPR int
}

// Make robot with default parameters or parameters from config file
func MakeDefaultDifferentialWheeledRobot() *DifferentialWheeledRobot {
	return &DifferentialWheeledRobot{
		BaseWidth:   config.ROBOT_BASE_WIDTH,
		WheelRadius: config.ROBOT_WHEEL_RADIUS,
		WheelRatio:  config.ROBOT_WHEEL_RATIO,
		OdometryPPR: config.ROBOT_ODOMETRY_PPR,
	}
}

// TypeString is "Differential Wheeled"
func (dwr *DifferentialWheeledRobot) TypeString() string {
	return "Differential Wheeled"
}

func (dwr *DifferentialWheeledRobot) GetParameters() interface{} {
	return dwr
}

// Calculate the distance the left and the right wheel will have to drive in
// order to accomplish a turn with a given radius and angle. The radius of the
// turn is given from the center of the axis, in left direction.
func (dwr *DifferentialWheeledRobot) TurnDistances(centerRadius, angle float64) (left, right float64) {
	radius := math.Abs(centerRadius) - dwr.BaseWidth/2
	short := radius * angle
	long := (radius + dwr.BaseWidth) * angle

	if centerRadius > 0 {
		return short, long
	}
	return long, short
}

// Calculate a new position based on an old, given the distance the left and
// the right wheel has rolled since the old position.
func (dwr *DifferentialWheeledRobot) RollPosition(distLeft, distRight float64, prev Position) Position {

	// Straight line
	if distLeft == distRight {
		return Position{
			prev.X + distLeft*math.Cos(prev.Theta),
			prev.Y + distLeft*math.Sin(prev.Theta),
			prev.Theta,
		}
	}

	// Turning
	turnRadius := dwr.BaseWidth * (distRight + distLeft) / (2 * (distRight - distLeft))
	angle := (distRight-distLeft)/dwr.BaseWidth + prev.Theta
	return Position{
		prev.X + turnRadius*(math.Sin(angle)-math.Sin(prev.Theta)),
		prev.Y - turnRadius*(math.Cos(angle)-math.Cos(prev.Theta)),
		angle,
	}

	// s := (distLeft + distRight) / 2.0
	// theta := (distRight-distLeft)/dwr.BaseWidth + prev.Theta
	// x := s*math.Cos(theta) + prev.X
	// y := s*math.Sin(theta) + prev.Y

	// return Position{x, y, theta}

}

// Calculate a new position base on an old, given the number of pulses (negative
// or positive) from each wheel since the last position.
func (dwr *DifferentialWheeledRobot) OdometryPosition(pulsesLeft, pulsesRight int, prev Position) Position {
	distancePerPulse := 2 * dwr.WheelRadius * math.Pi / float64(dwr.OdometryPPR)
	return dwr.RollPosition(distancePerPulse*float64(pulsesLeft), distancePerPulse*float64(pulsesRight), prev)
}
