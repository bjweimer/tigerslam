package astar

type Node struct {
	parent   *Node
	mapIndex int
	f        float64
	g        float64
}

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
	nq.nodes = a[0 : n-1]
	return node
}
