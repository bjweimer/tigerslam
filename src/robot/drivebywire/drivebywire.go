package drivebywire

import (
	"math"
	
	"robot/pathplanning/path"
)

type WireDriver struct {
	path *path.Path
}

// Get the Cross Track Error for the pose relative to the WireDriver's path.
func (wd *WireDriver) CrossTrackError(pose [3]float64) (float64, int) {
	
	// Loop through all the line segments in the path, find and return the
	// lowest distance, i.e. the minimum of the lowest distances from the point
	// to the individual line segments. This is the cross track error.
	min := math.MaxFloat64
	num := -1
	for i := 0; i < len(wd.path.Poses) - 1; i++ {
		dist, _ := minDistanceToLine(wd.path.Poses[i], wd.path.Poses[i + 1], pose)
		if dist < min {
			num = i
			min = dist
		}
	}
	return min, num
}

// Returns the minimum distance from pose p to line segment vw, as well as a
// boolean which is true if the projection of the point p falls within the line
// segment vw.
func minDistanceToLine(v3, w3 [3]float64, pose [3]float64) (float64, bool) {
	v := [2]float64{v3[0], v3[1]}
	w := [2]float64{w3[0], w3[1]}
	p := [2]float64{pose[0], pose[1]}
	dist2, inside := minDistanceToLineSquared(v, w, p)
	return math.Sqrt(dist2), inside
}

// Returns the square of the minimum distance from point p to the line segment
// vw, as well as a boolean which is true if the projection of the point p
// falls within the line segment vw.
func minDistanceToLineSquared(v, w [2]float64, p [2]float64) (float64, bool) {
	// l2 is the squared distance between v and w, i.e. |v-w|^2 (avoid sqrt)
	l2 := dist2(v, w)
	
	// v == w case
	if l2 == 0.0 {
		return dist2(p, v), false
	}
	
	// Consider the line extending the segment, parameterized as v + t (w - v).
	// We find projection of point p onto the line. It falls where
	// t = [(p-v) . (w-v)] / |w-v|^2
	t := dot([2]float64{p[0] - v[0], p[1] - v[1]}, [2]float64{w[0] - v[0], w[1] - v[1]}) / l2
	
	if t < 0 {
		// t < 0 implies we're beyond the 'v' end of the segment
		return dist2(p, v), false
	} else if t > 1 {
		// t > 1 implies we're beyond the 'w' end of the segment
		return dist2(p, w), false
	}
	
	// 0 < t < 1 implies the projection falls on the segment
	projection := [2]float64{
		v[0] + t * (w[0] - v[0]),
		v[1] + t * (w[1] - v[1]),
	}
	return dist2(p, projection), true
}

// Return the squared distance between the two points v and w, i.e. |w-v|^2.
func dist2(v, w [2]float64) float64 {
	return sqr(v[0] - w[0]) + sqr(v[1] - w[1])
}

// Square the number
func sqr(x float64) float64 {
	return x * x
}

// Dot product
func dot(v, w [2]float64) float64 {
	return v[0]*w[0] + v[1]*w[1]
}