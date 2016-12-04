package hector

import (
	"log"
	"math"
	//	"fmt"
	"image"

	"hectormapping"
	"hectormapping/datacontainer"
	"hectormapping/map/mapimages"
	"hectormapping/map/maprep"

	"robot/config"
	"robot/logging"
	"robot/model"
	"robot/sensors/lidar"
	"robot/sensors/odometry"
	"robot/sensors/sensor"
)

const TYPE_NAME = "hectorslam"

var logger *log.Logger

func init() {
	logger = logging.New()
}

// HectorSlam implements the Slam interface
type HectorSlam struct {
	hsp         *hectormapping.HectorSlamProcessor
	stopChan    chan bool
	lidarChan   chan sensor.SensorReading
	encoderChan chan sensor.SensorReading
	robot       *model.DifferentialWheeledRobot
	// leftPulses  int
	// rightPulses int
	filter            *OdomSlamEKF
	lastMapUpdatePose [3]float64
}

func MakeHectorSlam() *HectorSlam {
	slamProcessor := hectormapping.MakeHectorSlamProcessor(config.HECTORSLAM_GRIDMAP_RESOLUTION,
		config.HECTORSLAM_GRIDMAP_SIZE_X, config.HECTORSLAM_GRIDMAP_SIZE_Y,
		[2]float64{config.HECTORSLAM_GRIDMAP_START_X, config.HECTORSLAM_GRIDMAP_START_Y},
		config.HECTORSLAM_LEVELS)

	// Set update factors
	slamProcessor.SetUpdateFactorFree(config.HECTORSLAM_UPDATE_FACTOR_FREE)
	slamProcessor.SetUpdateFactorOccupied(config.HECTORSLAM_UPDATE_FACTOR_OCCUPIED)

	// Set minimum distance and angle for map update
	slamProcessor.SetMapUpdateMinDistDiff(config.HECTORSLAM_MAP_UPDATE_MIN_DIST_DIFF)
	slamProcessor.SetMapUpdateMinAngleDiff(config.HECTORSLAM_MAP_UPDATE_MIN_ANGLE_DIFF)

	return &HectorSlam{
		hsp:      slamProcessor,
		stopChan: make(chan bool),
		robot:    model.MakeDefaultDifferentialWheeledRobot(),
	}
}

func MakeHectorSlamFromMapRep(mapRep maprep.MapRepresentation) *HectorSlam {
	slamProcessor := hectormapping.MakeHectorSlamProcessorFromMapRep(mapRep)

	// Set update factors
	slamProcessor.SetUpdateFactorFree(config.HECTORSLAM_UPDATE_FACTOR_FREE)
	slamProcessor.SetUpdateFactorOccupied(config.HECTORSLAM_UPDATE_FACTOR_OCCUPIED)

	// Set minimum distance and angle for map update
	slamProcessor.SetMapUpdateMinDistDiff(config.HECTORSLAM_MAP_UPDATE_MIN_DIST_DIFF)
	slamProcessor.SetMapUpdateMinAngleDiff(config.HECTORSLAM_MAP_UPDATE_MIN_ANGLE_DIFF)

	return &HectorSlam{
		hsp:      slamProcessor,
		stopChan: make(chan bool),
		robot:    model.MakeDefaultDifferentialWheeledRobot(),
	}
}

func (hs *HectorSlam) Start() {

	// Start LIDAR sensor subscription
	hs.lidarChan = lidar.LidarSensor.Subscribe()

	// Start Odometry sensor subcription
	hs.encoderChan = odometry.OdometrySensor.Subscribe()

	// Start filter
	hs.filter = MakeOdomSlamEKF(hs.robot)
	hs.lastMapUpdatePose = [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}

	go hs.run()
}

func (hs *HectorSlam) run() {
	logger.Println("HectorSLAM now running")

	dataContainer := datacontainer.MakeDataContainer(config.LIDAR_NUM_DISTANCES)

	for {
		select {
		case <-hs.stopChan:

			// Stop
			hs.filter.Stop()
			return

		case sensorReading := <-hs.encoderChan:

			// If not using odometry, continue without update
			if !config.HECTORSLAM_USE_ODOMETRY {
				continue
			}

			// Update filter
			odometryReading, ok := sensorReading.(*odometry.OdometryReading)
			if !ok {
				logger.Println("Received invalid reading from encoder")
				continue
			}
			hs.filter.OdometryUpdate(odometryReading)

		case sensorReading := <-hs.lidarChan:

			// Run a SLAM update
			lidarReading, ok := sensorReading.(*lidar.LidarReading)
			if !ok {
				logger.Println("Received invalid reading from LIDAR")
				continue
			}

			// Obtain position estimate
			state := hs.filter.Estimate()

			// Make data container from LIDAR reading
			hs.LidarReadingToDataContainer(lidarReading, dataContainer, hs.hsp.GetScaleToMap())

			// Update SLAM
			hs.hsp.Update(dataContainer, [3]float64{state[0], state[1], state[2]})

			// Obtain position from SLAM
			matchedPos := hs.hsp.GetLastScanMatchPose()
			// matchedPos := hs.hsp.GetMapRepresentation().MatchData([3]float64{state[0], state[1], state[2]}, dataContainer, hs.hsp.GetLastScanMatchCovariance())

			// Update filter
			hs.filter.SLAMUpdate(matchedPos[0], matchedPos[1], matchedPos[2], lidarReading)

		}
	}

}

func (hs *HectorSlam) Stop() {
	// Stop LIDAR sensor subscription
	lidar.LidarSensor.Unsubscribe(hs.lidarChan)
	odometry.OdometrySensor.Unsubscribe(hs.encoderChan)

	// Stop the running loop
	hs.stopChan <- true

	logger.Println("HectorSLAM stopped.")
}

func (hs *HectorSlam) GetPosition() model.Position {
	// pose := hs.hsp.GetLastScanMatchPose()
	// return model.Position{pose[0], pose[1], pose[2]}
	if hs.filter == nil {
		return model.Position{}
	}

	state := hs.filter.Estimate()
	return model.Position{state[0], state[1], state[2]}
}

func (hs *HectorSlam) GetPositionHistory() []model.Position {
	return make([]model.Position, 0)
}

func (hs *HectorSlam) GetTypeName() string {
	return TYPE_NAME
}

// Warning: Assumes the map is quadratic
func (hs *HectorSlam) GetMapSizeMeters() float64 {
	mapDims := hs.hsp.GetGridMap().GetMapDimProperties()
	return mapDims.GetCellLength() * float64(mapDims.GetSizeX())
}

// Warning: Assumes the map is quadratic
func (hs *HectorSlam) GetMapSize() int {
	return hs.hsp.GetGridMap().GetMapDimProperties().GetSizeX()
}

func (hs *HectorSlam) GetMapImage() (image.Image, error) {
	return mapimages.GetMapImage(hs.hsp.GetMapRepresentation())
}

func (hs *HectorSlam) GetMapTile(zoomLevel uint, tileX, tileY int) (image.Image, error) {
	return mapimages.GetMapTile(hs.hsp.GetMapRepresentation(), zoomLevel, tileX, tileY)
}

func (hs *HectorSlam) GetOffsetX() float64 {
	return hs.hsp.GetGridMap().GetMapDimProperties().GetTopLeftOffset()[0]
}

func (hs *HectorSlam) GetOffsetY() float64 {
	return hs.hsp.GetGridMap().GetMapDimProperties().GetTopLeftOffset()[1]
}

func (hs *HectorSlam) GetMapRepresentation() maprep.MapRepresentation {
	return hs.hsp.GetMapRepresentation()
}

// Hector Mapping has it's own ideas of how the LIDAR data should be held. This
// function transforms a LidarReading into Hector Mapping's DataContainer,
// which means setting the data container's origo to (0, 0) and filling in
// values converted from the LidarReading's polar coordinate system (distances
// from the LIDAR with implicit angle) into a cartesian system, with (x, y)
// distance components from the origo. The function also does basic filtering,
// leaving out beams which should not be considered by the SLAM process.
func (hs *HectorSlam) LidarReadingToDataContainer(lidarReading *lidar.LidarReading,
	dataContainer *datacontainer.DataContainer, scaleToMap float64) {

	// Time the LIDAR uses on a sweep from the first measurement to the last
	// TODO: Should be a configurable parameter
	T_SPAN := 0.06667

	N := len(lidarReading.Distances)

	// Alpha is the angle of the first beam
	alpha := -lidarReading.Span / 2 * math.Pi / 180.0
	deltaAngle := lidarReading.Span / float64(N - 1) * math.Pi / 180
	angle := alpha

	// Configure dataContainer
	dataContainer.Clear()
	dataContainer.SetOrigo([2]float64{config.LIDAR_POSITION_X, config.LIDAR_POSITION_Y})

	// Get the estimated wheel speeds of the robot
	states := hs.filter.States()
	v_l := states[3]
	v_r := states[4]

	// Estimate theta_dot, the rate of rotation for the robot
	theta_dot := (v_r - v_l) / hs.robot.BaseWidth

	// Estimate speed of robot
	v := (v_r + v_l) / 2.0

	// Loop through all measurements, insert them into dataContainer
	for i := 0; i < N; i, angle = i + 1, angle + deltaAngle {

		// If distance is < 0.1 meters, disregard
		if lidarReading.Distances[i] < 100 {
			continue
		}

		// Distance in map
		dist := lidarReading.Distances[i] / 1000.0 * scaleToMap

		// Calculate naive distances, i.e. not taking delay between time of
		// measurement for the beam and time of scan into account.
		x_p := dist * math.Cos(angle)
		y_p := dist * math.Sin(angle)

		// Calculate the (negative) time delay
		d_i := - float64(N - i) / float64(N) * T_SPAN

		// Estimate differences in angle and position in robot y direction
		delta_theta_i := theta_dot * d_i
		delta_y_i := v * d_i

		// Corrigate measurements
		x_a := x_p * math.Cos(delta_theta_i) - y_p * math.Sin(delta_theta_i)
		y_a := x_p * math.Sin(delta_theta_i) + y_p * math.Cos(delta_theta_i) + delta_y_i

		// Add measurements to dataContainer
		if config.HECTORSLAM_USE_LIDAR_CORRECTION {
			// Corrected
			dataContainer.Add([2]float64{x_a, y_a})
		} else {
			// Un-corrected
			dataContainer.Add([2]float64{x_p, y_p})
		}
		
	}

}

// func LidarReadingToDataContainer(lidarReading *lidar.LidarReading,
// 	dataContainer *datacontainer.DataContainer, scaleToMap float64) {

// 	size := len(lidarReading.Distances)
// 	angle := -lidarReading.Span / 2 * math.Pi / 180.0
// 	deltaAngle := lidarReading.Span / float64(size-1) * math.Pi / 180.0

// 	dataContainer.Clear()
// 	dataContainer.SetOrigo([2]float64{config.LIDAR_POSITION_X, config.LIDAR_POSITION_Y})

// 	for i := 0; i < size; i++ {
// 		dist := lidarReading.Distances[i] / 1000.0
// 		if dist > 0.1 {
// 			dist *= scaleToMap
// 			meas := [2]float64{math.Cos(angle) * dist, math.Sin(angle) * dist}
// 			dataContainer.Add(meas)
// 		}

// 		angle += deltaAngle

// 	}
// }
