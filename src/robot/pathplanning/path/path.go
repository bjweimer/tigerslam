// Package path provides functionality for drive-by-wire paths.
//
// Paths are typically planned automatically or plotted by a user. They are a
// collection of waypoints in world (physical, metric) coordinates to be
// traversed in order. This package provides the data structure, as well as
// convenient methods.
package path

import (
	"math"
	"time"
)

// The path object consists of the poses of the path, together with
// an ID which can be used for identifying the path (e.g. see if it
// has changed.
type Path struct {
	Poses [][3]float64
	ID    int64
}

func MakePath() *Path {
	p := new(Path)
	p.Poses = make([][3]float64, 0)
	p.ID = time.Now().UnixNano()

	return p
}

// Simplify simplifies a path by deleting redundant information.
func (p *Path) Simplify() {
	p.simplifyStraightLines()
}

// simplifyStraightLines detects subsequent pairs of edges which have the exact
// same angle. The center node of such a edge pair can be deleted.
func (p *Path) simplifyStraightLines() {

	// Loop through all inner nodes of the path
	for i := 1; i < len(p.Poses)-1; {
		// Get backward and forward deltas
		bDeltaX, bDeltaY := deltas(p.Poses[i-1], p.Poses[i])
		fDeltaX, fDeltaY := deltas(p.Poses[i], p.Poses[i+1])

		// Get backward and forward angles
		bAngle := math.Atan2(bDeltaY, bDeltaX)
		fAngle := math.Atan2(fDeltaY, fDeltaX)

		if bAngle == fAngle {
			// The angles are the same, so the i'th node in the path is
			// redundant; remove it!
			p.Poses = append(p.Poses[:i], p.Poses[i+1:]...)
		} else {
			i++
		}
	}

}

// Smooth the path with a gradient descent interpolation algorithm, as seen at
// Udacity.com. Iteratively smoothens the path.
func (p *Path) Smooth(weightData, weightSmooth, tolerance float64) {

	// Create an qual set of poses - this will be the new poses of the path
	newPath := make([][3]float64, len(p.Poses))
	for i := range newPath {
		newPath[i] = p.Poses[i]
	}

	counter := 0
	change := tolerance
	for change >= tolerance {
		change = 0.0
		counter++

		// Iterate through all poses but the first and the last
		for i := 1; i < len(p.Poses)-1; i++ {
			for j := range p.Poses[0] {

				aux := newPath[i][j]
				newPath[i][j] += weightData * (p.Poses[i][j] - newPath[i][j])
				newPath[i][j] += weightSmooth * (newPath[i-1][j] + newPath[i+1][j] - 2.0*newPath[i][j])

				change += math.Abs(aux - newPath[i][j])
			}
		}
	}

	p.Poses = newPath

}

// Find deltaX and deltaY from two points
func deltas(a, b [3]float64) (deltaX, deltaY float64) {
	return b[0] - a[0], b[1] - a[1]
}
