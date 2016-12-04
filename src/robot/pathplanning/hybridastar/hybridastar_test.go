package hybridastar

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"testing"

	"robot/model"
	"robot/pathplanning/binarymap"
	"robot/pathplanning/path"
)

func TestBuildTransitions(t *testing.T) {

	hpp := &HybridPathPlanner{}
	hpp.robot = model.MakeDefaultDifferentialWheeledRobot()

	hpp.buildTransitions()

	// List the transitions
	for i, trans := range hpp.transitions {
		t.Logf("%d: %f, %f, %f: cost %f", i, trans.move[0], trans.move[1],
			trans.move[2], trans.cost)
	}

}

func TestMovePose(t *testing.T) {
	hpp := &HybridPathPlanner{}
	hpp.robot = model.MakeDefaultDifferentialWheeledRobot()

	p0 := [3]float64{5, 3, 0}
	move := [3]float64{1, 0, 0.7853981633974}
	p1 := hpp.movePose(p0, move)

	t.Logf("P1: %f, %f, %f", p1[0], p1[1], p1[2])
}

func TestGetMapIndex(t *testing.T) {
	hpp := &HybridPathPlanner{}
	hpp.binMap = binarymap.MakeOccGridMapBinary(0.1, [2]int{1024, 1024}, [2]float64{0, 0})

	i0 := hpp.getMapIndex([3]float64{0, 0, 0})
	t.Logf("Index of (0, 0): %d", i0)

	i1 := hpp.getMapIndex([3]float64{102.3, 102.3, 0})
	t.Logf("Index of (102.4, 102.4): %d", i1)
}

func TestMazePlanning(t *testing.T) {
	_testFreePlanning("maze",
		[3]float64{59, 50, 0},
		[3]float64{0, 0, 0},
		t)
}

func TestMapPlanning(t *testing.T) {
	_testFreePlanning("map",
		[3]float64{10, 10, 0},
		[3]float64{90, 90, 0},
		t)
}

func _testFreePlanning(mapName string, from, to [3]float64, t *testing.T) {
	hpp := &HybridPathPlanner{}
	hpp.binMap = BinaryMapFromPNG(mapName)
	hpp._maxMapIndex = hpp.binMap.GetSizeX()*hpp.binMap.GetSizeY() - 1
	hpp.robot = model.MakeDefaultDifferentialWheeledRobot()
	hpp.buildTransitions()
	hpp.radius = 1.0

	path, err := hpp.PlanPath(from, to)
	if err != nil {
		t.Fatal(err)
	}

	// Simplify path
	path.Simplify()

	LogPathToFile(path, "path-"+mapName)
}

func LogPathToFile(path *path.Path, filename string) {
	file, err := os.Create("testoutput/" + filename + ".csv")
	if err != nil {
		return
	}

	for _, pose := range path.Poses {
		file.Write([]byte(fmt.Sprintf("%f, %f, %f\n", pose[0], pose[1], pose[2])))
	}

	file.Close()
}

func BinaryMapFromPNG(filename string) *binarymap.OccGridMapBinary {
	file, err := os.Open("testinput/" + filename + ".png")
	if err != nil {
		return nil
	}

	// Decode the image
	img, _, err := image.Decode(file)
	size := [2]int{img.Bounds().Dx(), img.Bounds().Dy()}

	binMap := binarymap.MakeOccGridMapBinary(0.1, size, [2]float64{0, 0})

	for i := 0; i < size[0]; i++ {
		for j := 0; j < size[1]; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			cell := binMap.GetCell(i, j)
			if r == 0 && g == 0 && b == 0 {
				binMap.ConcreteGridFunctions.UpdateSetOccupied(cell)
			} else {
				binMap.ConcreteGridFunctions.UpdateSetFree(cell)
			}
		}
	}

	return binMap
}
