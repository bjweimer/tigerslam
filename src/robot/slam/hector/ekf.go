package hector

import (
	"math"
	"time"

	"github.com/skelterjohn/go.matrix"

	"robot/model"
	"robot/sensors/lidar"
	"robot/sensors/odometry"
)

type OdomSlamEKF struct {

	// Estimation-error covariance
	mP *matrix.DenseMatrix

	// State estimate
	mX *matrix.DenseMatrix

	// Design matrix Q
	mQ *matrix.DenseMatrix

	// Design matrix R
	mR *matrix.DenseMatrix

	robot *model.DifferentialWheeledRobot

	odomUpdateTime time.Time
	// lastOdomUpdateState matrix.Matrix

	slamUpdateTime time.Time
	// lastSlamUpdateState matrix.Matrix

	updateTime time.Time
}

func MakeOdomSlamEKF(robot *model.DifferentialWheeledRobot) *OdomSlamEKF {
	ekf := new(OdomSlamEKF)

	// ekf.lastOdomUpdateState = matrix.MakeDenseMatrix([]float64{0, 0, 0, 0, 0}, 5, 1)
	// ekf.lastSlamUpdateState = matrix.MakeDenseMatrix([]float64{0, 0, 0, 0, 0}, 5, 1)

	ekf.robot = robot

	ekf.mX = matrix.MakeDenseMatrix([]float64{0, 0, 0, 0, 0}, 5, 1)
	ekf.mP = matrix.Eye(5)

	ekf.mQ = matrix.Diagonal([]float64{2.0, 2.0, 2.0, 2.0, 2.0})
	ekf.mR = matrix.Diagonal([]float64{0.2, 0.2, 0.2, 0.2, 0.2})

	ekf.updateTime = time.Now()

	return ekf
}

func (o *OdomSlamEKF) Stop() {
	// Set v_l and v_r to 0, to stop propagation
	o.mX = matrix.Zeros(5, 1)
}

func (o *OdomSlamEKF) States() []float64 {
	return o.mX.Array()
}

// Given an odometry update d_l and d_r (distances left and right), produce a
// full measurement (x, y, theta, v_l, v_r), then update the filter.
func (o *OdomSlamEKF) OdometryUpdate(odometryReading *odometry.OdometryReading) {

	// If this is the first, disregard it but store the timestamp
	if o.odomUpdateTime.IsZero() {
		o.odomUpdateTime = odometryReading.GetTimestamp()
		return
	}

	// Find the two deltaTs: time since last filter update and since last
	// odometry update
	deltaTfilter := time.Since(o.updateTime)
	deltaTodom := time.Since(o.odomUpdateTime)

	// Make sure we update the timestamp for odometry
	defer func() { o.odomUpdateTime = odometryReading.GetTimestamp() }()

	// Compute distances d_l and d_r
	distancePerPulse := 2 * o.robot.WheelRadius * math.Pi / float64(o.robot.OdometryPPR)
	d_l := distancePerPulse * float64(odometryReading.LeftPulses)
	d_r := distancePerPulse * float64(odometryReading.RightPulses)

	// Compute the velocity we've had since last odometry update
	v_l := d_l / deltaTodom.Seconds()
	v_r := d_r / deltaTodom.Seconds()

	// Get position from last filter update
	prePos := model.Position{
		o.mX.Get(0, 0),
		o.mX.Get(1, 0),
		o.mX.Get(2, 0),
	}

	// Compute states x, y, theta through propagation with the distance we've
	// traveled since last filter update, with the speeds from odometry.
	newPos := o.robot.RollPosition(v_l*deltaTfilter.Seconds(), v_r*deltaTfilter.Seconds(), prePos)
	x := newPos.X
	y := newPos.Y
	theta := newPos.Theta

	// Create measurement vector
	mX := matrix.MakeDenseMatrix([]float64{x, y, theta, v_l, v_r}, 5, 1)

	// Update Kalman filter
	o.update(mX, odometryReading.GetTimestamp(), "ODOMETRY")

}

// LIDAR update produces a new estimate of (x, y, theta). Recover v_l and v_r,
// to get a full measurement (x, y, theta, v_l, v_r), then update the filter.
func (o *OdomSlamEKF) SLAMUpdate(x, y, theta float64, lidarReading *lidar.LidarReading) {

	// If this is the first, disregard it but store the timestamp
	if o.slamUpdateTime.IsZero() {
		o.slamUpdateTime = lidarReading.GetTimestamp()
		return
	}

	// Find the two deltaTs: the time since last filter update and since last
	// SLAM update
	// deltaTfilter := time.Since(o.updateTime)
	deltaTslam := time.Since(o.slamUpdateTime)
	defer func() { o.slamUpdateTime = lidarReading.GetTimestamp() }()

	// Obtain previous x, y, theta
	prevX := o.mX.Get(0, 0)
	prevY := o.mX.Get(1, 0)
	prevTheta := o.mX.Get(2, 0)

	// Assume motion consist of a rotation and a straight line
	// TODO: this assumes forward motion
	distance := math.Sqrt(math.Pow(x-prevX, 2) + math.Pow(y-prevY, 2))
	rotDistance := o.robot.BaseWidth / 2 * (theta - prevTheta)
	v_l := (distance + rotDistance) / deltaTslam.Seconds()
	v_r := (distance - rotDistance) / deltaTslam.Seconds()

	// Create measurement vector
	mX := matrix.MakeDenseMatrix([]float64{x, y, theta, v_l, v_r}, 5, 1)

	// Update Kalman Filter
	o.update(mX, lidarReading.GetTimestamp(), "SLAM")

}

func (o *OdomSlamEKF) dfdx(mX *matrix.DenseMatrix, delta_t time.Duration) *matrix.DenseMatrix {

	theta := mX.Get(2, 0)
	v_l := mX.Get(3, 0)
	v_r := mX.Get(4, 0)

	return matrix.MakeDenseMatrix([]float64{
		1, 0, -(v_r + v_l) / 2 * math.Sin(theta) * delta_t.Seconds(), math.Cos(theta) / 2 * delta_t.Seconds(), math.Cos(theta) / 2 * delta_t.Seconds(),
		0, 1, (v_r + v_l) / 2 * math.Cos(theta) * delta_t.Seconds(), math.Sin(theta) / 2 * delta_t.Seconds(), math.Sin(theta) / 2 * delta_t.Seconds(),
		0, 0, 1, 0, 0,
		0, 0, 0, 1, 0,
		0, 0, 0, 0, 1,
	}, 5, 5)

}

func (o *OdomSlamEKF) pMinus(mF, mPplus, mQ *matrix.DenseMatrix) *matrix.DenseMatrix {
	fp, _ := mF.TimesDense(mPplus)
	fpf, _ := fp.TimesDense(mF.Transpose())

	pMinus, _ := fpf.PlusDense(mQ)
	return pMinus
}

func (o *OdomSlamEKF) xMinus() *matrix.DenseMatrix {

	//mX = // Estimate from last update
	delta_t := time.Since(o.updateTime)

	pos := model.Position{
		o.mX.Get(0, 0),
		o.mX.Get(1, 0),
		o.mX.Get(2, 0),
	}

	v_l := o.mX.Get(3, 0)
	v_r := o.mX.Get(4, 0)

	newPos := o.robot.RollPosition(v_l*delta_t.Seconds(), v_r*delta_t.Seconds(), pos)

	return matrix.MakeDenseMatrix([]float64{newPos.X, newPos.Y, newPos.Theta, v_l, v_r}, 5, 1)
}

func (o *OdomSlamEKF) update(mY *matrix.DenseMatrix, timestamp time.Time, updateType string) {

	delta_t := time.Since(timestamp)

	mX := o.xMinus()
	mF := o.dfdx(mX, delta_t)
	mP := o.pMinus(mF, o.mP, o.mQ)

	mPplusmR, _ := mP.PlusDense(o.mR)
	mPplusmRinv, _ := mPplusmR.Inverse()
	mK, _ := mP.TimesDense(mPplusmRinv)

	if updateType == "ODOMETRY" {
		// Set upper three lines to zero
		mK.SetMatrix(0, 0, matrix.Zeros(5, 3))
		// mK.SetMatrix(0, 0, matrix.Zeros(3, 5))
	} else if updateType == "SLAM" {
		// Set bottom two lines to zero
		mK.SetMatrix(0, 3, matrix.Zeros(5, 2))
		// mK.SetMatrix(3, 0, matrix.Zeros(2, 5))
	} else {
		// Propagate only -- set all to zero
		mK = matrix.MakeDenseMatrix(make([]float64, 25), 5, 5)
	}

	mYminusmX, _ := mY.MinusDense(mX)
	mKtimesmYminusmX, _ := mK.TimesDense(mYminusmX)
	mXplus, _ := mX.PlusDense(mKtimesmYminusmX)

	IminusK, _ := matrix.Eye(5).MinusDense(mK)
	mPplus, _ := IminusK.TimesDense(mP)

	o.mX = mXplus
	o.mP = mPplus

	o.updateTime = timestamp
}

// Provide an estimate of the current state, based on propagation since last
// filter update.
func (o *OdomSlamEKF) Estimate() []float64 {
	x := o.xMinus()
	return x.Array()
}

// // Estimate the current position, given the time difference since the last
// // update and the wheel speeds.
// func (o *OdomSlamEKF) PositonEstimate() [3]float64 {

// 	// Get estimates of x, y, theta from last update

// 	return [3]float64{}
// }
