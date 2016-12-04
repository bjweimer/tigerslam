// Package pathfollowing provides algorithms for computing control signals for
// the motors suitable for making the robot follow a specific path.
package pathfollowing

import (
	"robot/pathplanning/path"
)

// Each path following algorithm implements a common interface, PathFollower.
type PathFollower interface {
	// Assigns a path to follow.
	SetPath(p path.Path)
	// Computes speeds for motors.
	SpeedUpdate(pos [3]float64) (left, right float64, finished bool)
}
