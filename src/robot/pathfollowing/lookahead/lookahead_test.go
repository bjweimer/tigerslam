package lookahead

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"testing"

	"robot/model"
	"robot/pathplanning/path"
)

var testPath = &path.Path{
	Poses: [][3]float64{
		[3]float64{0, 1, 0},
		[3]float64{6, 1, 0},
		[3]float64{6, 6, 0},
		[3]float64{1, 6, 0},
		[3]float64{1, 1, 0},
	},
}

func TestSegmentDistance(t *testing.T) {

	l := MakeLookahead()

	pointA := [2]float64{3, 3}
	pointB := [2]float64{7, 7}

	if got, want := l.segmentDistance(pointA, pointB), math.Sqrt(32); got != want {
		t.Errorf("Got %f, wanted %f", got, want)
	} else {
		t.Logf("Distance was %f", got)
	}

}

func TestCourseAngle(t *testing.T) {

	l := MakeLookahead()

	l.SetPath(testPath)

	pointA, pointB, _ := l.getPoints()

	courseAngle, s := l.courseAngle(pointA, pointB, [2]float64{7.5, 0.6})

	t.Logf("Course angle was %f", courseAngle)
	t.Logf("Along-track distance %f", s)

}

func TestSpeedUpdate(t *testing.T) {

	l := MakeLookahead()

	l.SetPath(testPath)

	v_l, v_r, finished := l.SpeedUpdate([3]float64{0, 0, 1.57})

	t.Logf("V_l: %f", v_l)
	t.Logf("V_r: %f", v_r)
	t.Logf("Finished: %t", finished)
}

// Proper testing: Assume some initial position, then use the robot model to
// propagate based on the speed updates, following the line.
func TestLookahead(t *testing.T) {

	l := MakeLookahead()
	l.SetPath(testPath)

	// Make the robot
	robot := model.MakeDefaultDifferentialWheeledRobot()

	// Start position
	pos := [3]float64{0, 0, 0}

	// Propagate for delta_t secs per update
	delta_t := 0.1

	// Get a CSV file for logging
	file, err := os.Create("testoutput/lookahead.csv")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	// Set up CSV
	writer := csv.NewWriter(file)

	for i := 0; ; i++ {

		// Print index
		// fmt.Printf("Iteration %d\n", i)

		// Get the speed update
		v_l, v_r, finished := l.SpeedUpdate(pos)

		// Distances
		dist_l := v_l * delta_t
		dist_r := v_r * delta_t

		// Propagate the robot
		position := robot.RollPosition(dist_l, dist_r, model.Position{pos[0], pos[1], pos[2]})
		pos[0] = position.X
		pos[1] = position.Y
		pos[2] = position.Theta

		// Output
		// t.Logf("i: %d, v_l: %f, v_r: %f, finished: %t", l.currIndex, v_l, v_r, finished)
		t.Logf("    Pos: %f, %f, %f", pos[0], pos[1], pos[2])

		// Write to CSV file
		writer.Write([]string{
			fmt.Sprintf("%f", pos[0]),
			fmt.Sprintf("%f", pos[1]),
			fmt.Sprintf("%f", pos[2]),
		})

		if finished {
			break
		}
	}

	writer.Flush()
}
