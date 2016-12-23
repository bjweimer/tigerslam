package astar

import (
	"container/heap"
	"errors"
	"fmt"
	"log"
	"math"

	"hectormapping/map/gridmap"

	"robot/config"
	"robot/logging"
	"robot/model"
	"robot/pathplanning/binarymap"
	"robot/pathplanning/path"
)

var logger *log.Logger

func init() {
	logger = logging.New()
}

type AstarPlanner struct {

	// Reference to the original grid map
	occMap gridmap.OccGridMap

	// Reference to an internal binary map
	binMap *binarymap.OccGridMapBinary

	robot *model.DifferentialWheeledRobot

	// Goal state in world coordinates
	goal           [3]float64
	goalIndex      int
	goalCellCoords [2]int

	// Start state in world coordinates
	start [3]float64

	// Maximum index of the binmap
	_maxMapIndex int

	// Neighbours, represented in the shift of index
	neighbours []int
}

// Create and initialize a AstarPlanner from a given gridmap. Create the
// internal map using a shrink factor, such that the number of cells is
// size(gridmap)/shrinkfactor.
func MakeAstarPlanner(gridMap gridmap.OccGridMap, robot *model.DifferentialWheeledRobot) *AstarPlanner {
	a := &AstarPlanner{
		robot:  robot,
		occMap: gridMap,
	}

	// Prepare an internal binary map representation
	a.binMap = binarymap.MakeShrunkenBinaryMap(gridMap, config.ASTAR_SHRINK_FACTOR, config.ASTAR_CHECK_RADIUS)
	a._maxMapIndex = a.binMap.GetSizeX()*a.binMap.GetSizeY() - 1

	// Build the neighbours
	a.neighbours = []int{
		-a.binMap.GetSizeX(), // North
		1,                    // East
		a.binMap.GetSizeX(),  // South
		-1,                   // West
	}

	return a
}

func (a *AstarPlanner) PlanPath(from, to [3]float64) (*path.Path, error) {

	// Set the new goal
	if err := a.setGoal(to); err != nil {
		return nil, err
	}

	// The set of nodes already evaluated
	closedMap := map[int]*Node{}

	// The set of tentative nodes to be evaluated, initially containing the
	// start node.
	openSet := MakeNodeQueue()
	openMap := map[int]bool{}

	// Add the start node
	openSet.Push(&Node{nil, a.getMapIndex(from), 0, 0})
	openSet.nodes[0].f = a.heuristic(openSet.nodes[0])
	openMap[openSet.nodes[0].mapIndex] = true

	var i int
	for i = 0; openSet.Len() > 0 && i < config.ASTAR_MAX_ITERATIONS; i++ {

		// The current node is the one from openSet with the lowest F value
		current := heap.Pop(openSet).(*Node)
		delete(openMap, current.mapIndex)

		// Check if current position is in the map
		if current.mapIndex < 0 || current.mapIndex > a._maxMapIndex {
			continue
		}

		// Check if we have now reached the goal node
		if current.mapIndex == a.goalIndex {
			logger.Printf("Successfully planned an A* path to the goal after %d iterations!\n", i)
			return a.reconstructPath(current), nil
		}

		// Add current to closedset
		closedMap[current.mapIndex] = current

		// For each neighbour
		for _, shift := range a.neighbours {

			// Create the (tentative) node, set up it's scores
			node := a.makeNode(current, shift)

			// Check if we already have a better value fo this mapindex
			if closedNode, ok := closedMap[node.mapIndex]; ok {
				if node.g >= closedNode.g {
					continue
				}
			}

			if _, ok := openMap[node.mapIndex]; !ok {
				heap.Push(openSet, node)
				openMap[node.mapIndex] = true
			}

		}
	}

	return nil, errors.New(fmt.Sprintf("No path found to goal. Used %d iterations.", i))
}

// Set a new goal
func (a *AstarPlanner) setGoal(goal [3]float64) error {
	goalPositionWorld := [2]float64{goal[0], goal[1]}
	goalPositionMap := a.binMap.GetMapCoords(goalPositionWorld)

	// Check if it's outside the map
	if a.binMap.PointOutOfMapBounds(goalPositionMap) {
		return errors.New("Goal out of map bounds")
	}

	// Check if it's occupied -- do it in the internal binary map
	if a.binMap.IsOccupied(int(goalPositionMap[0]), int(goalPositionMap[1])) {
		return errors.New("Goal is occupied")
	}

	a.goal = goal
	a.goalIndex = a.getMapIndex(goal)
	a.goalCellCoords = a.getCoords(a.goalIndex)

	//Debug: This converts world coords to map coords: fmt.Printf("Goalcoords: %d, %d\n", a.goalCellCoords[0], a.goalCellCoords[1])

	return nil
}

// Return the corresponding map index for the map pose
func (a *AstarPlanner) getMapIndex(pose [3]float64) int {
	mapPose := a.binMap.GetMapCoordsPose(pose)
	return int(mapPose[0]) + a.binMap.GetSizeX()*int(mapPose[1])
}

func (a *AstarPlanner) heuristic(node *Node) float64 {

	// Check if it's inside the map
	if node.mapIndex > a._maxMapIndex || node.mapIndex < 0 {
		return 1e9
	}

	// Check if it's occupied
	if a.binMap.GetCellByIndex(node.mapIndex).IsOccupied() {
		return 1e5
	}

	// Check the Euclidian distance to goal
	euclidian := a.euclidianDistanceToGoal(node.mapIndex)

	// Get the corresponding cell in the original map
	coords := a.getCoords(node.mapIndex)
	occMapCell := a.occMap.GetCell(coords[0]*config.ASTAR_SHRINK_FACTOR, coords[1]*config.ASTAR_SHRINK_FACTOR)
	if !occMapCell.IsFree() {

		// It's not free, so multiply by UNKNOWN_PUNISH
		euclidian *= config.ASTAR_UNKNOWN_PUNISH

	}

	return euclidian
}

// Get the world coordinates of the mapindex
func (a *AstarPlanner) getCoords(mapIndex int) [2]int {
	sizeX := a.binMap.GetSizeX()

	y := mapIndex / sizeX
	x := mapIndex % sizeX

	return [2]int{x, y}
}

// Calculate the euclidian distance to the goal from a map index
func (a *AstarPlanner) euclidianDistanceToGoal(mapIndex int) float64 {
	coords := a.getCoords(mapIndex)

	return math.Sqrt(math.Pow(float64(coords[0]-a.goalCellCoords[0]), 2) +
		math.Pow(float64(coords[1]-a.goalCellCoords[1]), 2))
}

// Make a node and set up it's values
func (a *AstarPlanner) makeNode(parent *Node, shift int) *Node {

	mapIndex := parent.mapIndex + shift
	g := parent.g + 1

	node := &Node{
		parent:   parent,
		mapIndex: mapIndex,
		g:        g,
	}

	node.f = node.g + a.heuristic(node)

	return node
}

func (a *AstarPlanner) reconstructPath(endNode *Node) *path.Path {

	p := path.MakePath()

	mdp := a.binMap.GetMapDimProperties()
	offset := mdp.GetTopLeftOffset()

	for current := endNode; current.parent != nil; current = current.parent {

		mapCoords := a.getCoords(current.mapIndex)
		p.Poses = append(p.Poses, [3]float64{
			float64(mapCoords[0])*a.binMap.GetCellLength() - offset[0],
			float64(mapCoords[1])*a.binMap.GetCellLength() - offset[1],
			0,
		})
	}

	// Reverse to get first node first
	length := len(p.Poses)
	rev := make([][3]float64, length)
	for i := range rev {
		rev[i] = p.Poses[length-i-1]
	}
	p.Poses = rev

	p.Smooth(config.ASTAR_SMOOTHING_DATA_WEIGHT, config.ASTAR_SMOOTHING_SMOOTH_WEIGHT, 0.001)
	p.Simplify()

	return p
}
