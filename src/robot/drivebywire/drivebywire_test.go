package drivebywire

import (
    "testing"
)

func TestMinDistanceToLine(t *testing.T) {
	v := [2]float64{3, 4}
	w := [2]float64{10, 4}
	
	var dist float64
	var inside bool
	
	dist, inside = minDistanceToLine(v, w, [3]float64{0, 0, 0})
	if dist != 5.0 || inside != false {
		t.Errorf("01 Dist was %f, inside was %t", dist, inside)
	}
	
	dist, inside = minDistanceToLine(v, w, [3]float64{5, 0, 0})
	if dist != 4.0 || inside != true {
		t.Errorf("02 Dist was %f, inside was %t", dist, inside)
	}
	
	dist, inside = minDistanceToLine(v, w, [3]float64{10, 4, 0})
	if dist != 0.0 || inside != true {
		t.Errorf("03 Dist was %f, inside was %t", dist, inside)
	}
	
	dist, inside = minDistanceToLine(v, w, [3]float64{11, 4, 0})
	if dist != 1.0 || inside != false {
		t.Errorf("04 Dist was %f, inside was %t", dist, inside)
	}
}

