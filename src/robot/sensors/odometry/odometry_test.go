package odometry

import (
	"testing"
)

func TestParseAnswer(t *testing.T) {

	answer := "H:23 V:-434;"

	e := MakeDefaultEncoder()
	left, right, err := e.parseAnswer(answer)

	if err != nil {
		t.Error(err)
	}

	if got, want := left, int64(-434); got != want {
		t.Logf("Got %d, wanted %d", got, want)
	}

	if got, want := right, int64(23); got != want {
		t.Logf("Got %d, wanted %d", got, want)
	}

	t.Logf("Interpreted as L%d, R%d", left, right)

}
