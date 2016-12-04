package los

import (
	"math"
	"testing"
)

const THRESHOLD = 1e-10

func floatApprox(f1, f2 float64) bool {
	if math.Abs(f1-f2) < THRESHOLD {
		return true
	}

	return false
}

type EpsilonTest struct {
	A, B, pos [2]float64
	s, e      float64
}

func TestEpsilon1(t *testing.T) {

	tests := map[string]*EpsilonTest{
		"Along X": &EpsilonTest{
			A: [2]float64{2, 5}, B: [2]float64{8, 5}, pos: [2]float64{7, 3},
			s: 5, e: 2,
		},
		"Test A": &EpsilonTest{
			A: [2]float64{3, -3}, B: [2]float64{6, 6}, pos: [2]float64{3, 4},
			s: 6.64078308635, e: -2.21359436212,
		},
		"Along Y": &EpsilonTest{
			A: [2]float64{2, 0}, B: [2]float64{2, 10}, pos: [2]float64{2, 9},
			s: 9, e: 0,
		},
		"Premature": &EpsilonTest{
			A: [2]float64{0, 0}, B: [2]float64{5, 0}, pos: [2]float64{-0.5, 0.7},
			s: -0.5, e: -0.7,
		},
	}

	for name, data := range tests {
		s, e := Epsilon(data.A, data.B, data.pos)

		if !floatApprox(s, data.s) {
			t.Errorf("%s: got s=%.11f (should be %f)", name, s, data.s)
		}

		if !floatApprox(e, data.e) {
			t.Errorf("%s: got e=%.11f (should be %f)", name, e, data.e)
		}

		if floatApprox(s, data.s) && floatApprox(e, data.e) {
			t.Logf("Test %s correct (s=%f, e=%f)", name, s, e)
		}
	}

}
