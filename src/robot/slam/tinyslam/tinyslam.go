// Package tinyslam implements a Go version of the TinySLAM algorithm.
//
// The TinySLAM algorithm was developed by Bruno Steux and Oussama El Hamzaoui,
// and more information is available e.g. at http://openslam.org/tinyslam.html.
// The algorithm seeks to be a small and simple implementation of SLAM, function
// on a simple gridmap where each cell is a integer value representing the
// likelihood that the cell is occupied.
package tinyslam

import (
	"fmt"
	"image"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"robot/config"
	"robot/logging"
	"robot/model"
	"robot/sensors/lidar"
	"robot/sensors/odometry"
	"robot/sensors/sensor"
	"robot/slam/tinyslam/gridmap"
	"robot/tools/intmath"
)

var logger *log.Logger

const TYPE_NAME = "tinyslam"

type Direction int

const (
	DIRECTION_FORWARD = iota
	DIRECTION_BACKWARD
	FINAL_MAP
)

//const NO_OBSTACLE gridmap.SimpleCell = 65500
//const OBSTACLE gridmap.SimpleCell = 0
const NO_OBSTACLE = 65500
const OBSTACLE = 0

func init() {
	logger = logging.New()
}

// TinySlam holds state information for the algorithm and implements the Slam
// interface.
type TinySlam struct {
	stopChan chan bool
	gridMap  *SlamSimpleMap
	robot    model.Robot

	// The instantaneous position, and log
	position        model.Position
	positionHistory []model.Position

	// Sensors
	lidar        *lidar.Lidar
	lidarChan    chan sensor.SensorReading
	encoder      *odometry.Encoder
	odometryChan chan sensor.SensorReading

	// Internal variables
	timestamp            time.Time
	direction            Direction
	thetadot             float64
	velocity             float64
	distance             float64
	montecarloIterations int

	// Config parameters
	sigmaXY    float64
	sigmaTheta float64
	holeWidth  int
}

// Local type so that we can extend
type SlamSimpleMap struct {
	gridmap.SimpleMap
}

// Cartesian Lidar Readings hold the raw reading from the LIDAR in addition
// to the corresponding X and Y points where the beams hit.
type cartesianLidarReading struct {
	lidar.LidarReading
	x     []float64
	y     []float64
	value []gridmap.SimpleCell
}

// Make a TinySLAM object
func MakeTinySlam(robot model.Robot) *TinySlam {
	// Initialize gridmap
	gridMap := &SlamSimpleMap{
		*gridmap.MakeSimpleMap(config.TINYSLAM_GRIDMAP_SIZE, config.TINYSLAM_GRIDMAP_RESOLUTION),
	}
	gridMap.Fill(gridmap.SimpleCell(NO_OBSTACLE * 0.50))

	// Initial position
	position := model.Position{
		gridMap.SizeMeters() / 2,
		gridMap.SizeMeters() / 2,
		0,
	}

	ts := &TinySlam{
		stopChan: make(chan bool),
		gridMap:  gridMap,
		robot:    robot,

		// Make space for many positions
		position:        position,
		positionHistory: make([]model.Position, 0, 10000),

		// Sensors
		lidar: lidar.LidarSensor,

		// Config parameters
		sigmaXY:              config.TINYSLAM_SIGMA_XY,
		sigmaTheta:           config.TINYSLAM_SIGMA_THETA,
		holeWidth:            config.TINYSLAM_HOLE_WIDTH,
		montecarloIterations: config.TINYSLAM_MONTECARLO_ITERATIONS,
	}

	return ts
}

// Return type name
func (ts *TinySlam) GetTypeName() string {
	return TYPE_NAME
}

// Start the SLAM progress
func (ts *TinySlam) Start() {
	// Start LIDAR sensor subscribtion
	ts.lidarChan = ts.lidar.Subscribe()
	go ts.run()
}

func (ts *TinySlam) GetOffsetX() float64 { return 0 }
func (ts *TinySlam) GetOffsetY() float64 { return 0 }

// Run is the running loop of the algorithm. It gathers sensor data and does
// map updates.
func (ts *TinySlam) run() {
	logger.Println("TinySLAM now running")
	for {
		select {
		case <-ts.stopChan:
			return
		case sensorReading := <-ts.lidarChan:
			lidarReading, ok := sensorReading.(*lidar.LidarReading)
			if !ok {
				fmt.Errorf("Received a sensor reading from lidar which was not a lidar reading")
			}

			ts.LidarMapBuilding(lidarReading, &ts.position)
		}
	}
}

// Stop the SLAM progress
func (ts *TinySlam) Stop() {
	// Stop LIDAR sensor subscription
	ts.lidar.Unsubscribe(ts.lidarChan)

	// Stop the running loop
	ts.stopChan <- true

	logger.Println("TinySLAM stopped.")
}

// Return the history of positions
func (ts *TinySlam) GetPositionHistory() []model.Position {
	return ts.positionHistory
}

// Return current position
func (ts *TinySlam) GetPosition() model.Position {
	return ts.position
}

// Return full map image
func (ts *TinySlam) GetMapImage() (image.Image, error) {
	return ts.gridMap.Image(0, 0, 0), nil
}

// Return map tile
func (ts *TinySlam) GetMapTile(zoomLevel uint, tileX, tileY int) (image.Image, error) {
	return ts.gridMap.ImageTile(zoomLevel, tileX, tileY)
}

// Size of map in meters
func (ts *TinySlam) GetMapSizeMeters() float64 {
	return ts.gridMap.SizeMeters()
}

// Size of map in cells
func (ts *TinySlam) GetMapSize() int {
	return ts.gridMap.Size()
}

// Construct a cartesian lidar reading from a raw lidar reading
func makeCartesianLidarReading(lidarReading lidar.LidarReading) *cartesianLidarReading {
	numDistances := len(lidarReading.Distances)

	r := &cartesianLidarReading{
		LidarReading: lidarReading,
		x:            make([]float64, numDistances),
		y:            make([]float64, numDistances),
		value:        make([]gridmap.SimpleCell, numDistances),
	}

	startAngle := -lidarReading.Span / 2 * math.Pi / 180.0
	deltaAngle := lidarReading.Span / float64(numDistances-1) * math.Pi / 180.0
	for i, angle := 0, startAngle; i < numDistances; i++ {
		if lidarReading.Distances[i] == 0 {
			r.x[i] = lidarReading.MaxDistance * math.Cos(angle)
			// ... (missing) correcting for speed
			r.y[i] = lidarReading.MaxDistance * math.Sin(angle)
			r.value[i] = NO_OBSTACLE
		} else {
			r.x[i] = float64(lidarReading.Distances[i]) / 1000 * math.Cos(angle)
			// ... (missing) correcting for speed
			r.y[i] = float64(lidarReading.Distances[i]) / 1000 * math.Sin(angle)
			r.value[i] = OBSTACLE
		}

		angle += deltaAngle
	}

	return r
}

// Calculate the "distance" (penalty value) from a scan to a map, based on a
// hypothetical position
func (sm *SlamSimpleMap) DistanceCartToMap(cart *cartesianLidarReading, pos model.Position) int {
	var sum int64 = 0
	var nbPoints int64 = 0

	x, y := 0, 0
	c := math.Cos(pos.Theta)
	s := math.Sin(pos.Theta)

	for i := range cart.Distances {
		if cart.value[i] != NO_OBSTACLE {
			x = int(sm.WorldToMapCoordinate(pos.X + c*cart.x[i] - s*cart.y[i]))
			y = int(sm.WorldToMapCoordinate(pos.Y + s*cart.x[i] + c*cart.y[i]))
			// Check boundaries
			if x >= 0 && x < sm.Size() && y >= 0 && y < sm.Size() {
				sum += int64(sm.At(x, y).(gridmap.SimpleCell))
				nbPoints++
			}
		}
	}

	if nbPoints > 0 {
		sum = sum * 1024 / nbPoints
	} else {
		sum = math.MaxInt32
	}

	return (int)(sum)
}

// Simply named "Map update, part 1" in TinySLAM paper. The function writes a
// laser ray to the map.
func (sm *SlamSimpleMap) MapLaserRay(x1, y1, x2, y2, xp, yp, alpha int, value int) {
	var x2c, y2c, dx, dy, dxc, dyc, x, err, errv, derrv, incerrv int
	var incv, sincv, incptrx, incptry, pixval, horiz, diago, ptroffset int
	var ptr *gridmap.SimpleCell

	size := sm.Size()

	// Check boundaries, about if we're outside the map
	if x1 < 0 || x1 >= size || y1 < 0 || y1 >= size {
		return
	}

	x2c = x2
	y2c = y2

	// Clipping
	if x2c < 0 {
		if x2c == x1 {
			return
		}
		y2c += (y2c - y1) * (-x2c) / (x2c - x1)
		x2c = 0
	}
	if x2c >= size {
		if x2c == x1 {
			return
		}
		y2c += (y2c - y1) * (size - 1 - x2c) / (x2c - x1)
		x2c = size - 1
	}
	if y2c < 0 {
		if y1 == y2c {
			return
		}
		x2c += (x1 - x2c) * (-y2c) / (y1 - y2c)
		y2c = 0
	}
	if y2c >= size {
		if y1 == y2c {
			return
		}
		x2c += (x1 - x2c) * (size - 1 - y2c) / (y1 - y2c)
		y2c = size - 1
	}

	dx = intmath.Abs(x2 - x1)
	dy = intmath.Abs(y2 - y1)
	dxc = intmath.Abs(x2c - x1)
	dyc = intmath.Abs(y2c - y1)

	if x2 > x1 {
		incptrx = 1
	} else {
		incptrx = -1
	}
	if y2 > y1 {
		incptry = size
	} else {
		incptry = -size
	}
	if value > NO_OBSTACLE {
		sincv = 1
	} else {
		sincv = -1
	}

	if dx > dy {
		derrv = intmath.Abs(xp - x2)
	} else {
		dx, dy = dy, dx
		dxc, dyc = dyc, dxc
		incptrx, incptry = incptry, incptrx
		derrv = intmath.Abs(yp - y2)
	}

	err = 2*dyc - dxc
	horiz = 2 * dyc
	diago = 2 * (dyc - dxc)
	errv = derrv / 2
	if derrv != 0 {
		incv = (value - NO_OBSTACLE) / derrv
	} else {
		incv = 0
	}
	incerrv = value - NO_OBSTACLE - derrv*incv

	ptroffset = y1*size + x1
	pixval = int(NO_OBSTACLE)

	for x = 0; x <= dxc; x, ptroffset = x+1, ptroffset+incptrx {
		ptr = &sm.Cells[ptroffset]

		if x > dx-int(2*derrv) {
			if x <= dx-int(derrv) {
				pixval += incv
				errv += incerrv
				if errv > derrv {
					pixval += sincv
					errv -= derrv
				}
			} else {
				pixval -= incv
				errv -= incerrv
				if errv < 0 {
					pixval -= sincv
					errv += derrv
				}
			}
		}

		// Integration into the map
		*ptr = (gridmap.SimpleCell)(((256-alpha)*int(*ptr) + alpha*pixval) >> 8)

		if err > 0 {
			ptroffset += incptry
			err += diago
		} else {
			err += horiz
		}
	}
}

// This function is simply called "Map update, part 2" in the TinySLAM paper. It
// Writes a cartesian reading to the map, given a position.
func (sm *SlamSimpleMap) MapUpdate(cart *cartesianLidarReading, pos *model.Position, quality, holeWidth int) {
	var x2p, y2p, dist, add float64
	var x1, y1, x2, y2, xp, yp, q int
	var value int

	c := math.Cos(pos.Theta)
	s := math.Sin(pos.Theta)

	x1 = int(sm.WorldToMapCoordinate(pos.X))
	y1 = int(sm.WorldToMapCoordinate(pos.Y))

	// Translate and rotate scan to robot position
	for i := range cart.x {
		x2p = c*cart.x[i] - s*cart.y[i]
		y2p = s*cart.x[i] + c*cart.y[i]

		xp = int(sm.WorldToMapCoordinate(pos.X + x2p))
		yp = int(sm.WorldToMapCoordinate(pos.Y + y2p))

		dist = math.Sqrt((float64)(x2p*x2p + y2p*y2p))
		add = float64(holeWidth) / 2 / dist / 1000.0

		x2p *= sm.WorldToMapCoordinate(1 + add)
		y2p *= sm.WorldToMapCoordinate(1 + add)
		x2 = int(sm.WorldToMapCoordinate(pos.X) + x2p)
		y2 = int(sm.WorldToMapCoordinate(pos.Y) + y2p)

		if cart.value[i] == NO_OBSTACLE {
			q = quality / 4
			value = NO_OBSTACLE
		} else {
			q = quality
			value = OBSTACLE
		}
		sm.MapLaserRay(x1, y1, x2, y2, xp, yp, q, value)
	}
}

// Monte Carlo Search for the best position of a cartesian reading, given an
// approximated position. The output is an estimated improved position, fitting
// the map better. The implementation is simplistic.
func (sm *SlamSimpleMap) monteCarloSearch(cart *cartesianLidarReading,
	start_pos *model.Position, sigma_xy, sigma_theta float64, stop int,
	bd *int) model.Position {

	var currentpos, bestpos, lastbestpos model.Position
	var currentdist, bestdist, lastbestdist int
	counter := 0
	debug := 0

	if stop < 0 {
		debug = 1
		stop = -stop
	}

	currentpos, bestpos, lastbestpos = *start_pos, *start_pos, *start_pos
	currentdist = sm.DistanceCartToMap(cart, currentpos)
	bestdist, lastbestdist = currentdist, currentdist

	for counter < stop {
		currentpos = lastbestpos
		currentpos.X = rand.NormFloat64()*sigma_xy + currentpos.X
		currentpos.Y = rand.NormFloat64()*sigma_xy + currentpos.Y
		currentpos.Theta = rand.NormFloat64()*sigma_theta + currentpos.Theta

		currentdist = sm.DistanceCartToMap(cart, currentpos)

		if currentdist < bestdist {
			bestdist = currentdist
			bestpos = currentpos
			if debug != 0 {
				//				fmt.Printf("Monte carlo ! %f %f %f %d (count = %d)\n", bestpos.x, bestpos.y, bestpos.theta, bestdist, counter)
			}
		}

		counter++

		if counter > stop/3 {
			if bestdist < lastbestdist {
				lastbestpos = bestpos
				lastbestdist = bestdist
				counter = 0
				sigma_xy *= 0.5
				sigma_theta *= 0.5
			}
		}
	}

	if bd != nil {
		*bd = bestdist
	}

	return bestpos
}

// Same as monteCarloSearch, but capable of utilizing more CPU cores.
func (sm *SlamSimpleMap) concMonteCarloSearch(cart *cartesianLidarReading, start_pos *model.Position,
	sigma_xy, sigma_theta float64, stop int) model.Position {

	// Acquire the number of CPUs (cores)
	n := runtime.NumCPU()
	positionChan := make(chan model.Position)
	stopPerRoutine := int(float64(stop) * 0.5)

	// Start n go routines, each calculating a "best position"
	for i := 0; i < n; i++ {
		go func() {
			positionChan <- sm.monteCarloSearch(cart, start_pos, sigma_xy, sigma_theta, stopPerRoutine, nil)
		}()
	}

	bestpos := *start_pos
	bestdist := sm.DistanceCartToMap(cart, bestpos)

	// Retrieve the positions, find the best of the best
	for i := 0; i < n; i++ {
		position := <-positionChan
		distance := sm.DistanceCartToMap(cart, position)

		if distance < bestdist {
			bestpos = position
			bestdist = distance
		}
	}

	return bestpos
}

// Take a LidarReading, correct the state position based on this, and integrate
// the lidar data into the map. The method assumes an estimated position.
func (ts *TinySlam) LidarMapBuilding(lidarReading *lidar.LidarReading, estPosition *model.Position) {
	// Correct the state position
	cartReading := makeCartesianLidarReading(*lidarReading)
	//	correctedPosition := ts.gridMap.monteCarloSearch(cartReading, estPosition,
	//		ts.sigmaXY, ts.sigmaTheta, ts.montecarloIterations, nil)
	correctedPosition := ts.gridMap.concMonteCarloSearch(cartReading, estPosition,
		ts.sigmaXY, ts.sigmaTheta, ts.montecarloIterations)

	// Integrate to the map
	ts.gridMap.MapUpdate(cartReading, &correctedPosition, 50, ts.holeWidth)

	// Delta position is the difference in position since previous map building
	deltaPosition := model.Position{
		correctedPosition.X - ts.position.X,
		correctedPosition.Y - ts.position.Y,
		correctedPosition.Theta - ts.position.Theta,
	}
	deltaDistance := math.Sqrt(deltaPosition.X*deltaPosition.X + deltaPosition.Y*deltaPosition.Y)
	ts.distance += deltaDistance

	// Update velocity, thetadot
	if !ts.timestamp.IsZero() {
		deltaTime := lidarReading.GetTimestamp().Sub(ts.timestamp)
		ts.velocity = deltaDistance / deltaTime.Seconds()
		ts.thetadot = deltaPosition.Theta / deltaTime.Seconds()
	}
	ts.timestamp = lidarReading.GetTimestamp()

	// Update slam position
	ts.position = correctedPosition
}
