package occbase

import (
	"math"

	"robot/tools/intmath"

	"github.com/skelterjohn/go.matrix"

	"hectormapping/datacontainer"
	"hectormapping/map/gridmap"
	"hectormapping/map/gridmap/base"
	"hectormapping/utils"
)

type OccGridMapBase struct {
	base.GridMapBase
	ConcreteGridFunctions gridmap.GridFunctions
	currUpdateIndex       int
	currMarkOccIndex      int
	currMarkFreeIndex     int
}

func MakeOccGridMapBase(mapResolution float64, size [2]int, offset [2]float64, cellExample gridmap.Cell) *OccGridMapBase {
	return &OccGridMapBase{
		GridMapBase:       *base.MakeGridMapBase(mapResolution, size, offset, cellExample),
		currUpdateIndex:   0,
		currMarkOccIndex:  -1,
		currMarkFreeIndex: -1,
	}
}

func (ogmb *OccGridMapBase) UpdateSetOccupied(index int) {
	ogmb.ConcreteGridFunctions.UpdateSetOccupied(ogmb.GetCellByIndex(index))
}

func (ogmb *OccGridMapBase) UpdateSetFree(index int) {
	ogmb.ConcreteGridFunctions.UpdateSetFree(ogmb.GetCellByIndex(index))
}

func (ogmb *OccGridMapBase) UpdateUnsetFree(index int) {
	ogmb.ConcreteGridFunctions.UpdateUnsetFree(ogmb.GetCellByIndex(index))
}

func (ogmb *OccGridMapBase) GetGridProbabilityMap(xMap, yMap int) float64 {
	return ogmb.ConcreteGridFunctions.GetGridProbability(ogmb.GetCell(xMap, yMap))
}

func (ogmb *OccGridMapBase) GetGridProbabilityMapByIndex(index int) float64 {
	return ogmb.ConcreteGridFunctions.GetGridProbability(ogmb.GetCellByIndex(index))
}

func (ogmb *OccGridMapBase) IsOccupied(xMap, yMap int) bool {
	return ogmb.GetCell(xMap, yMap).IsOccupied()
}

func (ogmb *OccGridMapBase) IsFree(xMap, yMap int) bool {
	return ogmb.GetCell(xMap, yMap).IsFree()
}

func (ogmb *OccGridMapBase) IsOccupiedByIndex(index int) bool {
	return ogmb.GetCellByIndex(index).IsOccupied()
}

func (ogmb *OccGridMapBase) IsFreeByIndex(index int) bool {
	return ogmb.GetCellByIndex(index).IsFree()
}

func (ogmb *OccGridMapBase) GetObstacleThreshold() float64 {
	temp := ogmb.GetCellExample().Copy()
	temp.ResetGridCell()
	return ogmb.ConcreteGridFunctions.GetGridProbability(temp)
}

func (ogmb *OccGridMapBase) SetUpdateFreeFactor(factor float64) {
	ogmb.ConcreteGridFunctions.SetUpdateFreeFactor(factor)
}

func (ogmb *OccGridMapBase) SetUpdateOccupiedFactor(factor float64) {
	ogmb.ConcreteGridFunctions.SetUpdateOccupiedFactor(factor)
}

// Updates the map using the given scan data and robot pose
// @param dataContainer Contains the laser scan data
// @param robotPoseWorld The 2D robot pose in world coordinates
func (ogmb *OccGridMapBase) UpdateByScan(dataContainer *datacontainer.DataContainer, robotPoseWorld [3]float64) {

	ogmb.currMarkFreeIndex = ogmb.currUpdateIndex + 1
	ogmb.currMarkOccIndex = ogmb.currUpdateIndex + 2

	// Get pose in map coordinates from pose in world coordinates
	mapPose := ogmb.GetMapCoordsPose(robotPoseWorld)

	// Get a 2D homogenous transform that can be left-multiplied to a robot
	// coordinates vector to get world coordinates of that vector
	poseTransform := utils.TransformationMatrix2D(mapPose)

	// Get start point of all laser beams in map coordinates (same for all
	// beams, store in robot coords in dataContainer).
	origo := dataContainer.GetOrigo()
	origo[0], origo[1] = origo[0]/ogmb.GetCellLength(), origo[1]/ogmb.GetCellLength()
	temp, err := poseTransform.TimesDense(matrix.MakeDenseMatrix(append(origo[:], 1), 3, 1))
	if err != nil {
		panic(err)
	}
	scanBeginMapf := temp.Array()

	// get the integer vector of laser beams start point
	scanBeginMapi := [2]int{int(scanBeginMapf[0] + 0.5), int(scanBeginMapf[1] + 0.5)}

	// Get the number of valid beams in current scan
	numValidElems := dataContainer.GetSize()

	// Iterate over all valid laser beams
	for i := 0; i < numValidElems; i++ {

		// Get map coordinates of current beam endpoint
		vecEntry := dataContainer.GetVecEntry(i)
		t, err := poseTransform.TimesDense(matrix.MakeDenseMatrix(append(vecEntry[:], 1), 3, 1))
		if err != nil {
			panic(err)
		}
		scanEndMapf := t.Array()

		// Get integer map coordinates of current beam endpoint
		scanEndMapi := [2]int{int(scanEndMapf[0] + 0.5), int(scanEndMapf[1] + 0.5)}

		// Update map using a bresenham variant for drawing line from beam
		// start to beam endpoint in map coordinates.
		if scanBeginMapi != scanEndMapi {
			ogmb.UpdateLineBresenhami(scanBeginMapi, scanEndMapi, 0)
		}
	}

	// Tell the map that it has been updated
	ogmb.SetUpdated()

	// Increase update index (used for updating grid cells only once per
	// incoming scan).
	ogmb.currUpdateIndex += 3
}

// Update a line ranging from beginMap to endMap (in map integer coordinates)
// using the Bresenham2D algortihm. Draw the line only if both beginMap and
// endMap are within the map bounds. Also mark the end point of the line as
// occupied.
func (ogmb *OccGridMapBase) UpdateLineBresenhami(beginMap, endMap [2]int, maxLength uint64) {

	if maxLength <= 0 {
		maxLength = math.MaxUint64
	}

	x0 := beginMap[0]
	y0 := beginMap[1]

	// Check if beam start point is inside the map, cancel update if this is
	// not the case.
	if (x0 < 0) || (x0 >= ogmb.GetSizeX()) || (y0 < 0) || (y0 >= ogmb.GetSizeY()) {
		//	if !ogmb.HasGridValue(x0, y0) {
		//		logger.Printf("BeginMap is not on the map: (%d, %d)\n", x0, y0)
		return
	}

	x1 := endMap[0]
	y1 := endMap[1]

	// Check if beam end point is inside the map, cancel pdate if this is not
	// the case.
	if (x1 < 0) || (x1 >= ogmb.GetSizeX()) || (y1 < 0) || (y1 >= ogmb.GetSizeY()) {
		//	if !ogmb.HasGridValue(x1, y1) {
		//		logger.Printf("EndMap is not on the map: (%d, %d)\n", x1, y1)
		return
	}

	dx := x1 - x0
	dy := y1 - y0

	abs_dx := intmath.Abs(dx)
	abs_dy := intmath.Abs(dy)

	offset_dx := utils.Sign(dx)
	offset_dy := utils.Sign(dy) * ogmb.GetSizeX()

	startOffset := beginMap[1]*ogmb.GetSizeX() + beginMap[0]

	if abs_dx >= abs_dy {
		// X is dominant
		error_y := abs_dx / 2
		ogmb.Bresenham2D(abs_dx, abs_dy, error_y, offset_dx, offset_dy, startOffset)
	} else {
		// Y is dominant
		error_x := abs_dy / 2
		ogmb.Bresenham2D(abs_dy, abs_dx, error_x, offset_dy, offset_dx, startOffset)
	}

	endOffset := endMap[1]*ogmb.GetSizeX() + endMap[0]
	ogmb.BresenhamCellOcc(endOffset)
}

// Mark a cell as free
func (ogmb *OccGridMapBase) BresenHamCellFree(offset int) {
	cell := ogmb.GetCellByIndex(offset)

	if cell.GetUpdateIndex() < ogmb.currMarkFreeIndex {
		ogmb.ConcreteGridFunctions.UpdateSetFree(cell)
		cell.SetUpdateIndex(ogmb.currMarkFreeIndex)
	}
}

// Mark a cell as occupied
func (ogmb *OccGridMapBase) BresenhamCellOcc(offset int) {
	cell := ogmb.GetCellByIndex(offset)

	if cell.GetUpdateIndex() < ogmb.currMarkOccIndex {

		// If this cell has been updated as free in the current iteration,
		// revert this.
		if cell.GetUpdateIndex() == ogmb.currMarkFreeIndex {
			ogmb.ConcreteGridFunctions.UpdateUnsetFree(cell)
		}

		ogmb.ConcreteGridFunctions.UpdateSetOccupied(cell)
		cell.SetUpdateIndex(ogmb.currMarkOccIndex)
	}
}

// Draw a line using the Bresenham algorithm. See other sources for details of
// the Bresenham line drawing algorithm (e.g. Wikipedia). The algorithm is
// common in drawing approximations of straight lines at angles in a pixel
// (grid) map/image.
func (ogmb *OccGridMapBase) Bresenham2D(abs_da, abs_db, error_b, offset_a, offset_b, offset int) {
	ogmb.BresenHamCellFree(offset)

	end := abs_da - 1

	for i := 0; i < end; i++ {
		offset += offset_a
		error_b += abs_db

		if error_b >= abs_da {
			offset += offset_b
			error_b -= abs_da
		}

		ogmb.BresenHamCellFree(offset)
	}
}
