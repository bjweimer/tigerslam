// Package lookahead implements Lookahead-based steering, as described in
// The Handbook of Marine Craft Hydrodynamics and Motion Control.
//
// Lookahead steering is a simple steering mechanism, where the course angle
// is separated into two parts, as chi_d(e) = chi_p + chi_r(e), where
// chi_p = alpha_k for the current line segment. The mechanism considers pairs
// of waypoints as a line segment k.
package lookahead

import (
	"errors"
	// "fmt"
	"log"
	"math"

	"hectormapping/utils"

	"robot/config"
	"robot/logging"
	"robot/pathfollowing/los"
	"robot/pathplanning/path"
)

var logger *log.Logger

// Lookahead implements the PathFollower interface. It takes a path, and
// gives control signals.
type Lookahead struct {
	// The path we're currently following
	path *path.Path
	// Used to track the
	currIndex int
}

func init() {
	logger = logging.New()
}

// Makes a plain Lookahead object (no path set).
func MakeLookahead() *Lookahead {
	return &Lookahead{}
}

// Set the path for the Lookahead object. Should only be done once for a path.
// The object will assume it's near the first point of the path.
func (l *Lookahead) SetPath(p *path.Path) {
	l.path = p
	l.currIndex = 0
}

// Get control signals given some position. If the path traversion is
// finished, the flag "finished" will be set. Otherwise, the left and right
// return float64s will assume some values [-1, 1].
func (l *Lookahead) SpeedUpdate(pos [3]float64) (left, right float64, finished bool) {
	k := 0.5 // Set the scale constant for motor PID during turns and course corrections - see line # 93 below
	// Get the points. If this fails, we're finished.
	pointA, pointB, err := l.getPoints()
	if err != nil {
		return 0, 0, true
	} else {
		finished = false
	}

	//logger.Printf("Current Position: X = %.3v  Y = %.3v  Theta = %.3v\n", pos[0], pos[1], pos[2])

	// Get course angle and along-track distance
	chi_d, s := l.courseAngle(pointA, pointB, [2]float64{pos[0], pos[1]})

	// If we're past the distance of the line segment, move on to next one
	// (takes effect on next update)
	if s > l.segmentDistance(pointA, pointB)-config.LOOKAHEAD_DISTANCE/2 {
		l.currIndex++
		//	logger.Printf("Current Line Segment Index = %v, e_chi = %v\n", l.currIndex, chi_d-pos[2])
	}

	// We now have the desired course angle chi_d, and can compare this to the
	// angle of the current position in order to produce a control update.
	// Compute the error angle.
	e_chi := utils.NormalizeAngle(chi_d - pos[2])

	//logger.Printf("Index = %v and e_chi = %.3v\n", l.currIndex, e_chi)

	// BJW CODE: Limit e_chi to -1.57 to +1.57 range (i.e. -/+ 90 degrees).
	if e_chi < -1.57 {
		e_chi = -1.57
	}

	if e_chi > 1.57 {
		e_chi = 1.57
	}

	// This lets the robot go faster on straightaways ***********************************************************
	if math.Abs(e_chi) < 0.05 {
		k = 1.0
	}

	// This didn't work: k = (-50/157)*math.Abs(e_chi) + 1.0

	// BJW CODE: If e_chi is negative, turn clockwise; otherwise turn counterclockwise.
	if e_chi <= 0 {
		//logger.Printf("Clockwise: Index = %v and e_chi = %.3v\n", l.currIndex, e_chi)
		left = k * (((25.0 / 157.0) * e_chi) + 0.5)
		right = k * (((75.0 / 157.0) * e_chi) + 0.5)
		//logger.Printf("left = %v and right = %v\n\n", left, right)
	} else {
		//logger.Printf("Counterclockwise: Index = %v and e_chi = %.3v\n", l.currIndex, e_chi)
		left = k * (((-75.0 / 157.0) * e_chi) + 0.5)
		right = k * (((-25.0 / 157.0) * e_chi) + 0.5)
		//logger.Printf("left = %v and right = %v\n\n", left, right)
	}

	// We can now caluclate a "delta_rl", i.e. a difference between the
	// control signals to right and left wheels (right - left), which should
	// be used, as simply p*e_chi. P is the proportional factor in a PID
	// controller.
	// ORIGINAL CODE: delta_rl := config.LOOKAHEAD_P * e_chi

	//logger.Println("delta_rl = ", delta_rl)

	// Split the delta into two parts, add speed
	// ORIGINAL CODE: right = 0.5*delta_rl + config.LOOKAHEAD_U
	// ORIGINAL CODE: left = -0.5*delta_rl + config.LOOKAHEAD_U

	return

}

// Compute the course angle and the along-track distance
func (l *Lookahead) courseAngle(pointA, pointB, pos [2]float64) (chi_d, s float64) {

	// Alpha is the angle between the line segment and the x axis.
	alpha := los.Alpha(pointA, pointB)

	// chi_p is the first part of the angle we want, and represents the
	// path tangential angle.
	chi_p := alpha

	// Get the lookahead distance delta
	delta := config.LOOKAHEAD_DISTANCE

	// Get s, and e, the along-track distance and cross-track error
	s, e := los.Epsilon(pointA, pointB, pos)

	// fmt.Printf("e=%f, pointA: %f, %f, pointB: %f, %f\n", e, pointA[0], pointA[1], pointB[0], pointB[1])

	// Compute the velocity-path relative angle chi_r, which ensures that the
	// velocity is directed toward a point on the path that is located a
	// lookahead distance delta > 0 ahead of the direct projection of pos on
	// to the path segment.
	chi_r := math.Atan(-e / delta)

	chi_d = chi_p - chi_r

	// fmt.Printf("Chi_p=%f, Chi_r=%f, Chi_d=%f\n", chi_p, chi_r, chi_d)

	return
}

// Returns the two points defining the currently used line segment of the
// path. This will start off as the two first points of the path, then
// propagate to the next two and so on. This function assumes the index is
// valid, i.e. that it is not too high.
func (l *Lookahead) getPoints() (pointA, pointB [2]float64, err error) {

	if l.currIndex > len(l.path.Poses)-2 || l.currIndex < 0 {
		err = errors.New("Invalid index.")
		return
	}

	poseA := l.path.Poses[l.currIndex]
	pointA = [2]float64{poseA[0], poseA[1]}

	poseB := l.path.Poses[l.currIndex+1]
	pointB = [2]float64{poseB[0], poseB[1]}

	return
}

// Get the distance between two points, i.e. the maximum distance we should
// travel along the particular line segment.
func (l *Lookahead) segmentDistance(pointA, pointB [2]float64) float64 {
	return math.Sqrt(math.Pow(pointB[0]-pointA[0], 2) + math.Pow(pointB[1]-pointA[1], 2))
}
