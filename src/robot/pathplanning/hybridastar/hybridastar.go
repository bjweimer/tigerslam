// Hybrid A* is inspired by the "Hybrid A*" described in Montemerlo et. al.
// "Junior: The Stanford Entry in the Urban Challenge". It uses a (discrete)
// grid map, but uses a robot motion model to search along paths that are
// realizible (continuous paths). In each search step, it searches along a
// discrete (finite) set of possible continuous motions.
package hybridastar

import (
	"container/heap"
	"errors"
	"fmt"
	"math"

	"image/png"
	"os"

	"hectormapping/map/gridmap"
	"hectormapping/map/mapimages"

	"robot/model"
	"robot/pathplanning/binarymap"
	"robot/pathplanning/path"
)

const UNKNOWN_PUNISH = 100.0

// A transition consists of the relative coordinates it results in, and a cost.
// A transition should be executable for the robot.
type Transition struct {
	move [3]float64
	cost float64
}

// For simplicity, a combine functions in a HybridPathPlanner.
type HybridPathPlanner struct {
	// Holds a set of allowable transitions (movements) the robot can make.
	transitions []*Transition
	// Model of the robot
	robot *model.DifferentialWheeledRobot
	// Reference to the original grid map
	occMap gridmap.OccGridMap
	// Reference to an internal binary map
	binMap *binarymap.OccGridMapBinary
	// Goal state in world coordinates
	goal [3]float64
	// Start state in world coordinates
	start [3]float64
	// Radius of the goal
	radius float64

	// Maximum index of the binMap
	_maxMapIndex int
}

// Create and initialize a HybridPathPlanner from a given gridMap and a robot
// model. Create the internal map using a shrinkFactor, such that the number
// of cells is gridMap/shrinkFactor. The radius represents the radius of a
// circle around the goal position which should be considered allowable.
func MakeHybridPathPlanner(gridMap gridmap.OccGridMap, robot *model.DifferentialWheeledRobot, shrinkFactor int, radius float64) *HybridPathPlanner {
	hpp := &HybridPathPlanner{
		robot:  robot,
		occMap: gridMap,
		radius: radius,
	}

	// Prepare an internal binary map representation
	// hpp.binMap = binarymap.BinaryMapFromOccGridMap(gridMap, shrinkFactor)
	hpp.binMap = binarymap.MakeShrunkenBinaryMap(gridMap, shrinkFactor, radius)
	hpp._maxMapIndex = hpp.binMap.GetSizeX()*hpp.binMap.GetSizeY() - 1

	// DEBUG -- save an image of the binmap
	img, _ := mapimages.GetGridmapImage(hpp.binMap)
	if img != nil {
		f, _ := os.Create("binmap.png")
		png.Encode(f, img)
		f.Close()
	}

	// Build the set of allowable transitions
	hpp.buildTransitions()

	return hpp
}

func (hpp *HybridPathPlanner) PlanPath(from, to [3]float64) (*path.Path, error) {

	fmt.Printf("From: %.2f, %.2f, %.2f\n", from[0], from[1], from[2])
	fmt.Printf("To: %.2f, %.2f, %.2f\n", to[0], to[1], to[2])

	if err := hpp.setGoal(to); err != nil {
		return nil, err
	}

	counter := 0

	// The set of nodes already evaluated
	closedMap := map[int]*Node{}

	// The set of tentative nodes to be evalueated, initially containing the
	// start node.
	openSet := MakeNodeQueue()
	openMap := map[int]bool{}

	// Add the start node
	openSet.Push(&Node{nil, from, hpp.getMapIndex(from), 0, 0})
	openSet.nodes[0].f = hpp.heuristic(openSet.nodes[0])
	openMap[openSet.nodes[0].mapIndex] = true

	// Continue until the openSet is empty
	for openSet.Len() > 0 {

		// The current node is the one from openSet with the lowest F value
		current := heap.Pop(openSet).(*Node)
		delete(openMap, current.mapIndex)

		counter++

		// Check if position is in the map
		currMapPos := hpp.binMap.GetMapCoords([2]float64{current.position[0], current.position[1]})
		if hpp.binMap.PointOutOfMapBounds(currMapPos) {
			continue
		}

		// Check if we have now reached the goal node
		if hpp.goalIsReached(current.position) {
			fmt.Printf("Checked %d nodes.\n", counter)
			return reconstructPath(current), nil
		}

		// Add current to closedset
		closedMap[current.mapIndex] = current

		// For each neighbour, implemented by a transition
		for _, trans := range hpp.transitions {

			// Create the (tentative) node, set up it's scores
			node := hpp.makeNode(current, trans)

			// Check if we already have a better value for this mapindex
			if closedNode, ok := closedMap[node.mapIndex]; ok && node.g >= closedNode.g {
				continue
			}

			// If transition node not in openset or tentative g score < gscore[neighbour]
			if _, ok := openMap[node.mapIndex]; !ok {
				heap.Push(openSet, node)
				openMap[node.mapIndex] = true
			}
		}
	}

	return nil, errors.New("Path not found.")
}

// Heuristic
func (hpp *HybridPathPlanner) heuristic(node *Node) float64 {

	// Check if it's inside the map
	if node.mapIndex > hpp._maxMapIndex || node.mapIndex < 0 {
		return 1e9
	}

	// Check if it's occupied
	if hpp.binMap.GetCellByIndex(node.mapIndex).IsOccupied() {
		return 1e5
	}

	// Check the Euclidian distance to goal
	euclidian := hpp.euclidianDistanceToGoal(node.position)

	// Get the corresponding cell in the original map
	occMapCoords := hpp.occMap.GetMapCoords([2]float64{node.position[0], node.position[1]})
	occMapCell := hpp.occMap.GetCell(int(occMapCoords[0]), int(occMapCoords[1]))
	if !occMapCell.IsFree() {

		// It's not free, so multiply by UNKNOWN_PUNISH
		euclidian *= UNKNOWN_PUNISH

	}

	return euclidian
}

// Determines if a goal is reached.
func (hpp *HybridPathPlanner) goalIsReached(pose [3]float64) bool {

	// Calculate the euclidian distance between the pose and the goal
	distance := hpp.euclidianDistanceToGoal(pose)

	// See if it's inside the radius
	if distance < hpp.radius {
		return true
	}

	return false
}

// Set a new goal
func (hpp *HybridPathPlanner) setGoal(goal [3]float64) error {
	goalPositionWorld := [2]float64{goal[0], goal[1]}
	goalPositionMap := hpp.binMap.GetMapCoords(goalPositionWorld)

	fmt.Printf("World(%.2f, %.2f) = Map(%.2f, %.2f)\n", goalPositionWorld[0], goalPositionWorld[1], goalPositionMap[0], goalPositionMap[1])

	// Check if it's outside the map
	if hpp.binMap.PointOutOfMapBounds(goalPositionMap) {
		return errors.New("hpp.SetGoal: Goal out of map bounds.")
	}

	// Check if it's occupied -- do it in the internal binary map
	if hpp.binMap.IsOccupied(int(goalPositionMap[0]), int(goalPositionMap[1])) {
		return errors.New("Goal is occupied.")
	}

	hpp.goal = goal

	return nil
}

// Build a list of allowable transitions for the robot.
func (hpp *HybridPathPlanner) buildTransitions() {
	if hpp.robot == nil {
		return
	}

	hpp.transitions = make([]*Transition, 0)
	startPos := model.Position{}
	var endPos model.Position

	// Turn left, forward, turn right, reverse
	rolls := [][3]float64{
		[3]float64{0.3, 0.5, 0.6},
		[3]float64{0.5, 0.5, 0.5},
		[3]float64{0.5, 0.3, 0.6},
		[3]float64{-0.5, -0.5, 10.0},
	}

	//	var left, right float64
	//	rolls := make([][3]float64, 3)
	//	left, right = hpp.robot.TurnDistances(0.2, 0.7853981633974) // pi/4
	//	rolls[0] = [3]float64{left, right, (left + right) / 2}
	//	rolls[1] = [3]float64{0.7, 0.7, 0.7}
	//	left, right = hpp.robot.TurnDistances(-0.2, 0.7853981633974) // pi/4
	//	rolls[2] = [3]float64{left, right, (left + right) / 2}

	//	rolls := [][3]float64{
	//		[3]float64{-0.2, 0.1, 0.45},
	//		[3]float64{0.5, 0.5, 0.5},
	//		[3]float64{2.0, 2.0, 2.0},
	//		[3]float64{0.1, -0.2, 0.45},
	//	}

	// Smooth + hard
	//	rolls := [][3]float64{
	//		[3]float64{-0.2, 0.1, 0.5},
	//		[3]float64{0.4, 0.5, 0.45},
	//		[3]float64{0.5, 0.5, 0.5},
	//		[3]float64{0.5, 0.4, 0.45},
	//		[3]float64{0.1, -0.2, 0.5},
	//	}

	// Smooth
	//	rolls := [][3]float64{
	//		[3]float64{0.4, 0.5, 0.45},
	//		[3]float64{0.5, 0.5, 0.5},
	//		[3]float64{0.5, 0.4, 0.45},
	//	}

	for _, roll := range rolls {
		endPos = hpp.robot.RollPosition(roll[0], roll[1], startPos)
		transition := &Transition{[3]float64{endPos.X, endPos.Y, endPos.Theta}, roll[2]}
		hpp.transitions = append(hpp.transitions, transition)
	}
}

// Calculate the euclidian distance to the goal
func (hpp *HybridPathPlanner) euclidianDistanceToGoal(pose [3]float64) float64 {
	return math.Sqrt(math.Pow(pose[0]-hpp.goal[0], 2) +
		math.Pow(pose[1]-hpp.goal[1], 2))
}

// Make a node from a parent node and a transition. Should update position, g
// score, map index and f score (with heuristic), as well as set the parent
// node as the node's parent for back-tracing.
func (hpp *HybridPathPlanner) makeNode(parent *Node, transition *Transition) *Node {
	position := hpp.movePose(parent.position, transition.move)
	g := parent.g + transition.cost

	node := &Node{
		parent:   parent,
		position: position,
		mapIndex: hpp.getMapIndex(position),
		g:        g,
	}

	node.f = g + hpp.heuristic(node)

	return node
}

// Return the resulting pose from standing in oldPose and moving move, both
// specified by (x, y, theta)
func (hpp *HybridPathPlanner) movePose(pose, move [3]float64) [3]float64 {
	return [3]float64{
		move[0]*math.Cos(pose[2]) - move[1]*math.Sin(pose[2]) + pose[0],
		move[0]*math.Sin(pose[2]) + move[1]*math.Cos(pose[2]) + pose[1],
		pose[2] + move[2],
	}
}

// Return the corresponding map index for the map pose
func (hpp *HybridPathPlanner) getMapIndex(pose [3]float64) int {
	mapPose := hpp.binMap.GetMapCoordsPose(pose)
	return int(mapPose[0]) + hpp.binMap.GetSizeX()*int(mapPose[1])
}

// Check if a node with map index mapIndex is in the NodeQueue
func (hpp *HybridPathPlanner) mapIndexInQueue(set *NodeQueue, mapIndex int) bool {
	for i := range set.nodes {
		if set.nodes[i].mapIndex == mapIndex {
			return true
		}
	}
	return false
}

func reconstructPath(endNode *Node) *path.Path {
	p := path.MakePath()

	for current := endNode; current.parent != nil; current = current.parent {
		p.Poses = append(p.Poses, current.position)
	}

	// Reverse to get first node first
	length := len(p.Poses)
	rev := make([][3]float64, length)
	for i := range rev {
		rev[i] = p.Poses[length-i-1]
	}
	p.Poses = rev

	return p
}
