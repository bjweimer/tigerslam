package driver

import (
	"testing"
)

func speedFormatTestHelper(speed int, want string, t *testing.T) {
	if got := speedFormat(speed); got != want {
		t.Errorf("Got %s, wanted %s.", got, want)
	} else {
		t.Logf("Got %s, wanted %s.", got, want)
	}
}

func TestSpeedFormat(t *testing.T) {

	speedFormatTestHelper(0, "0000", t)
	speedFormatTestHelper(15, "0015", t)
	speedFormatTestHelper(2000, "2000", t)

}

func TestDefaultMotor(t *testing.T) {
	m := MakeDefaultMotor()

	testString := "balla"

	// Connect
	defer m.Disconnect()
	err := m.Connect()
	if err != nil {
		t.Error(err)
	}

	// Write it
	err = m.write(testString)
	if err != nil {
		t.Error(err)
	}

	// Read it
	retrieved, err := m.read()
	if err != nil {
		t.Error(err)
	}

	if testString != retrieved {
		t.Errorf("Sent %s, got %s back", testString, retrieved)
	}
}

func TestSetSpeedsDecimal(t *testing.T) {
	m := MakeDefaultMotor()

	// Connect
	err := m.Connect()
	defer m.Disconnect()
	if err != nil {
		t.Fatal(err)
	}

	err = m.SetSpeedsDecimal(1.0, -1.0)
	if err != nil {
		t.Error(err)
	}

	if got, want := m.speed_left, RANGE; got != want {
		t.Errorf("Left speed is %d, want %d", got, want)
	} else {
		t.Logf("Left speed is %d, want %d", got, want)
	}

	if got, want := m.speed_right, -RANGE; got != want {
		t.Errorf("Right speed is %d, want %d", got, want)
	} else {
		t.Logf("Left speed is %d, want %d", got, want)
	}
}
