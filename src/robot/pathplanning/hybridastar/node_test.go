package hybridastar

import (
    "testing"
    "container/heap"
)

func TestNodeQueue(t *testing.T) {
	nodeQueue := MakeNodeQueue()
	
	heap.Push(nodeQueue, &Node{f: 2})
	heap.Push(nodeQueue, &Node{f: 1})
	heap.Push(nodeQueue, &Node{f: 3})
	heap.Push(nodeQueue, &Node{f: 0.3})
	
	// Take the nodes out again, should arrive lowest-highest f value
	last := -1.0
	for _ = range nodeQueue.nodes {
		current := heap.Pop(nodeQueue).(*Node)
		t.Logf("Arrived: %f", current.f)
		if current.f < last {
			t.Error("Did not arrive in ascending order")
		}
		last = current.f
	}
}

