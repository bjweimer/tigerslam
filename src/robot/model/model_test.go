package model

import (
    "testing"
    "math"
)

var robot = &DifferentialWheeledRobot{0.3, 0.05, 1, 200}

func TestTurnDistances(t *testing.T) {
	left, right := robot.TurnDistances(-0.85, math.Pi)
	t.Logf("Left: %f, Right: %f", left, right)
}

func TestRollPosition(t *testing.T) {
	
	// Start in origin
	position := Position{0, 0, 0}
	t.Log(position)
	
	// Straight line
	position = robot.RollPosition(2, 2, position)
	t.Log(position)
	
	// Turn 90 degrees to the left
	position = robot.RollPosition(- robot.BaseWidth * math.Pi / 4, robot.BaseWidth * math.Pi / 4, position)
	t.Log(position)
	
	// Do a full semi circle back to origin, facing negative y direction
	position = robot.RollPosition(math.Pi - robot.BaseWidth / 2 * math.Pi, math.Pi + robot.BaseWidth / 2 * math.Pi, position)
	t.Log(position)
	
	// Turn 90 degrees to the left
	position = robot.RollPosition(- robot.BaseWidth * math.Pi / 4, robot.BaseWidth * math.Pi / 4, position)
	t.Log(position)
	
	tol := 1e-15
	if math.Abs(position.X) > tol || math.Abs(position.Y) > tol || math.Abs(position.Theta - 2 * math.Pi) > tol {
		t.Errorf("Position is: %s", position)
	}
}

func TestOdometryPosition(t *testing.T) {
	
	// Start in origin
	position := Position{0, 0, 0}
	t.Log(position)
	
	// Drive in a straight line
	position = robot.OdometryPosition(robot.OdometryPPR, robot.OdometryPPR, position)
	t.Log(position)
	
	if math.Abs(position.X - 2 * math.Pi * robot.WheelRadius) > 1e-15 {
		t.Error("Position is: %s", position)
	}
	
}