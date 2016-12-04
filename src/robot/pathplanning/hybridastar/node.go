package hybridastar

import (
)

// Node for the A* algorithm
type Node struct {
	parent *Node
	position [3]float64
	mapIndex int // The index (cell identifier) of the position in the map
	f float64 // Estimated total cost from start to goal through this node
	g float64 // Cost from start along the best known path
}

// NodeQueue is a priority queue for nodes, implemented as a heap supported by
// the container/heap package.
type NodeQueue struct {
	nodes []*Node
}

func MakeNodeQueue() *NodeQueue {
	return &NodeQueue{
		nodes: make([]*Node, 0),
	}
}

// Len is the number of elements in the collection.
func (nq *NodeQueue) Len() int {
	return len(nq.nodes)
}

// Less returns whether the element with index i should sort before the element
// with index j. 
func (nq *NodeQueue) Less(i, j int) bool {
	return nq.nodes[i].f < nq.nodes[j].f
}

// Swap swaps the elements with indexes i and j.
func (nq *NodeQueue) Swap(i, j int) {
	nq.nodes[i], nq.nodes[j] = nq.nodes[j], nq.nodes[i]
}

// Push
func (nq *NodeQueue) Push(x interface{}) {
	nq.nodes = append(nq.nodes, x.(*Node))
}

// Pop
func (nq *NodeQueue) Pop() interface{} {
	a := nq.nodes
	n := len(a)
	node := a[n-1]
	nq.nodes = a[0 : n - 1]
	return node
}

