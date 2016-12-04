// Package los provides functions common for several path following
// algorithms, inspired by The Handbook of Marine Craft Hydrodynamics and
// Motion Control, chapter 10.
package los

import (
	"math"
)

// Computes the angle between the line segment A-B and the X axis.
func Alpha(pointA, pointB [2]float64) (alpha float64) {
	return math.Atan2(pointB[1]-pointA[1], pointB[0]-pointA[0])
}

// Calculates along-track distance (tangential to path) and cross-track error
// (normal to path) from two path waypoints and a position. The along-track
// distance (s) is the distance the position corresponds to along the path
// segment. The cross track error is the position's displacement perpendicular
// to the path segment (distance from line). Details are explained in The
// Handbook of Marine Craft Hydrodynamics and Motion Control, around page 258.
// Path following algorithms will need to minimize e, lim[t->inf] e(t) = 0.
func Epsilon(pointA, pointB [2]float64, pos [2]float64) (s, e float64) {

	// Consider a path-fixed reference frame with origin in pointA whose x
	// axis has been rotated by a positive angle alpha:
	alpha := Alpha(pointA, pointB)

	// Cache sin and cos values
	cosAlpha := math.Cos(alpha)
	sinAlpha := math.Sin(alpha)

	// Compute along-track distance (tangential to path)
	s = (pos[0]-pointA[0])*cosAlpha + (pos[1]-pointA[1])*sinAlpha

	// Compute cross-track error (normal to path)
	e = -(pos[0]-pointA[0])*sinAlpha + (pos[1]-pointA[1])*cosAlpha
	e = -e

	return
}
