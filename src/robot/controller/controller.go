package controller

import (
	"errors"

	//	"robot/config"
	"robot/model"
	"robot/motor"
	"robot/sensors"
	"robot/slam"
)

// State holds all state information about the whole robot. The object pointer
// can be passed around to different modules, which in turn can manipulate
// the state through methods.
type Controller struct {
	SlamController   *slam.SlamController
	SensorController *sensors.SensorController
	MotorController  *motor.MotorController
	Robot            model.Robot
}

// Makes an arbitrary state. Under default operation use MakeDefaultState().
func MakeController(robot model.Robot, sensors *sensors.SensorController) *Controller {
	diffWheeledRobot := robot.(*model.DifferentialWheeledRobot)

	return &Controller{
		SlamController:   slam.MakeSlamController(),
		MotorController:  motor.MakeMotorController(diffWheeledRobot),
		Robot:            robot,
		SensorController: sensors,
	}
}

// Make a state from default parameters and parameters found in the config
// module. This is the one to use in i.e. the program's Main module.
func MakeDefaultController() *Controller {
	robot := model.MakeDefaultDifferentialWheeledRobot()
	sensors := sensors.MakeDefaultSensorController()

	return MakeController(robot, sensors)
}

func (c *Controller) PlanPath(goal [3]float64) error {

	// Get map from slam
	slam := c.SlamController.GetSlam()
	if slam == nil {
		return errors.New("Slam not initialized")
	}
	occMap := slam.GetMapRepresentation().GetGridMap(0)

	// Get current position
	position := slam.GetPosition()

	return c.MotorController.PlanPath(occMap, [3]float64{position.X, position.Y, position.Theta}, goal)

}

func (c *Controller) FollowPath() error {

	slam := c.SlamController.GetSlam()
	if slam == nil {
		return errors.New("Slam not initialized")
	}

	c.MotorController.FollowPath(slam)

	return nil
}
