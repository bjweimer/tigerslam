// Package pathplanning provides utilities for planning a path from one place
// in a OccGridMap to another, without touching an obstructed area.
package pathplanning

import (
	"robot/pathplanning/path"
)

type PathPlanner interface {
	PlanPath(from, to [3]float64) (*path.Path, error)
}
