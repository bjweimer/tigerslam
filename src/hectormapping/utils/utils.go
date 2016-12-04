package utils

import (
	"math"

	"github.com/skelterjohn/go.matrix"
)

// Makes a 2D (2x2) transformation matrix from the transformation vector, i.e.
// rotation and scaling
func TransformationMatrix2D(transVector [3]float64) *matrix.DenseMatrix {

	transformation := matrix.Eye(3)

	sin := math.Sin(transVector[2])
	cos := math.Cos(transVector[2])

	transformation.Set(0, 0, cos)
	transformation.Set(1, 0, sin)
	transformation.Set(0, 1, -sin)
	transformation.Set(1, 1, cos)

	transformation.Set(0, 2, transVector[0])
	transformation.Set(1, 2, transVector[1])

	return transformation
}

// Normalize angle pos
func NormalizeAnglePos(angle float64) float64 {
	return angle
	// return math.Mod(math.Mod(angle, 2.0*math.Pi)+2.0*math.Pi, 2.0*math.Pi)
}

// Normalize angle
func NormalizeAngle(angle float64) float64 {
	return angle
	// a := NormalizeAnglePos(angle)
	// if a > math.Pi {
	// 	a -= 2.0 * math.Pi
	// }
	// return a
}

func PoseDifferenceLargerThan(pose1, pose2 [3]float64, distanceDiffThresh, angleDiffThresh float64) bool {

	if math.Sqrt((pose1[0]-pose2[0])*(pose1[0]-pose2[0])+(pose1[1]-pose2[1])*(pose1[1]-pose2[1])) > distanceDiffThresh {
		return true
	}

	angleDiff := pose1[2] - pose2[2]

	if angleDiff > math.Pi {
		angleDiff -= 2.0 * math.Pi
	} else if angleDiff < -math.Pi {
		angleDiff += 2.0 * math.Pi
	}

	if math.Abs(angleDiff) > angleDiffThresh {
		return true
	}

	return false

}

// Sign
// Note that this function is not the regular which has {1, 0, -1} as output,
// but a version which hector uses, which only has {1, -1} as output.
func Sign(x int) int {
	if x > 0 {
		return 1
	}
	return -1
}
