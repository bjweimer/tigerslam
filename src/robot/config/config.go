// Provides configuration variables to the entire program, based on the goconf
// package. Configuration can be done in some configuration file. The package
// guarantees that the configuration variables exist. The variables may or may
// not change value during execution, and it is up to the user package to judge
// whether this should be considered.
package config

import (
	"fmt"
	//	"path/filepath"
	"code.google.com/p/goconf/conf"
)

// Where to look for config files
var configFilePaths = []string{
	"",
	"assets\\config\\",
	"assets/config/",
	"src/assets/config/",
	"..\\assets\\config\\",
	"../assets/config/",
	"C:\\Users\\mikaelbe\\work\\tigerslam\\assets\\config\\", // Ugly, needed for testing
}

const section = "default"

var ConfigFile *conf.ConfigFile

func init() {
	var err error

	// Load configuration file
	for i := range configFilePaths {
		//		fmt.Println(filepath.Abs(configFilePaths[i] + "config.cfg"))
		ConfigFile, err = conf.ReadConfigFile(configFilePaths[i] + "config.cfg")
		if err == nil {
			fmt.Printf("Using config file: %s\n", configFilePaths[i]+"config.cfg")
			break
		}
	}
	if err != nil {
		panic(err)
	}

	// Fill all values -- This should ideally be done in a nicer fashion
	// which still has variables as package exports, not a single variable
	// slice or object, for easier code. Maybe something like this could have
	// been done: https://gist.github.com/1307606
	ASSETS_ROOT = getString(section, "assets_root")

	API_URL = getString(section, "api_url")
	STATIC_URL = getString(section, "static_url")
	STATIC_ROOT = ASSETS_ROOT + getString(section, "static_root")
	TEMPLATE_ROOT = ASSETS_ROOT + getString(section, "template_root")
	BASE_TEMPLATE = getString(section, "base_template")
	WEB_ADDRESS = getString(section, "web_address")
	REALM = getString(section, "realm")
	USERNAME = getString(section, "username")
	PASSPHRASE = getString(section, "passphrase")
	LOG_UPDATE_RATE = getInt(section, "log_update_rate")

	DATASETS_ROOT = ASSETS_ROOT + getString(section, "datasets_root")

	SENSORLOGS_ROOT = ASSETS_ROOT + getString(section, "sensorlogs_root")
	SENSORLOGS_TIME_FORMAT = getString(section, "sensorlogs_time_format")
	USE_LIDAR = getBool(section, "use_lidar")
	LIDAR_COM_NAME = getString(section, "lidar_com_name")
	LIDAR_BAUD_RATE = getInt(section, "lidar_baud_rate")
	LIDAR_NUM_DISTANCES = getInt(section, "lidar_num_distances")
	LIDAR_RADIAL_SPAN = getFloat64(section, "lidar_radial_span")
	LIDAR_MAX_DISTANCE = getFloat64(section, "lidar_max_distance")
	LIDAR_POSITION_X = getFloat64(section, "lidar_position_x")
	LIDAR_POSITION_Y = getFloat64(section, "lidar_position_y")

	USE_ODOMETRY = getBool(section, "use_odometry")
	ODOMETRY_COM_NAME = getString(section, "odometry_com_name")
	ODOMETRY_BAUD_RATE = getInt(section, "odometry_baud_rate")

	MAX_SPEED = getFloat64(section, "max_speed")

	ROBOT_BASE_WIDTH = getFloat64(section, "robot_base_width")
	ROBOT_WHEEL_RADIUS = getFloat64(section, "robot_wheel_radius")
	ROBOT_WHEEL_RATIO = getFloat64(section, "robot_wheel_ratio")
	ROBOT_ODOMETRY_PPR = getInt(section, "robot_odometry_ppr")

	SLAM_ALGORITHM = getString(section, "slam_algorithm")

	TINYSLAM_GRIDMAP_SIZE = getInt(section, "tinyslam_gridmap_size")
	TINYSLAM_GRIDMAP_RESOLUTION = getInt(section, "tinyslam_gridmap_resolution")
	TINYSLAM_SIGMA_XY = getFloat64(section, "tinyslam_sigma_xy")
	TINYSLAM_SIGMA_THETA = getFloat64(section, "tinyslam_sigma_theta")
	TINYSLAM_HOLE_WIDTH = getInt(section, "tinyslam_hole_width")
	TINYSLAM_MONTECARLO_ITERATIONS = getInt(section, "tinyslam_montecarlo_iterations")

	HECTORSLAM_GRIDMAP_SIZE_X = getInt(section, "hectorslam_gridmap_size_x")
	HECTORSLAM_GRIDMAP_SIZE_Y = getInt(section, "hectorslam_gridmap_size_y")
	HECTORSLAM_GRIDMAP_RESOLUTION = getFloat64(section, "hectorslam_gridmap_resolution")
	HECTORSLAM_GRIDMAP_START_X = getFloat64(section, "hectorslam_gridmap_start_x")
	HECTORSLAM_GRIDMAP_START_Y = getFloat64(section, "hectorslam_gridmap_start_y")
	HECTORSLAM_LEVELS = getInt(section, "hectorslam_levels")
	HECTORSLAM_UPDATE_FACTOR_FREE = getFloat64(section, "hectorslam_update_factor_free")
	HECTORSLAM_UPDATE_FACTOR_OCCUPIED = getFloat64(section, "hectorslam_update_factor_occupied")
	HECTORSLAM_MAP_UPDATE_MIN_ANGLE_DIFF = getFloat64(section, "hectorslam_map_update_min_angle_diff")
	HECTORSLAM_MAP_UPDATE_MIN_DIST_DIFF = getFloat64(section, "hectorslam_map_update_min_dist_diff")
	HECTORSLAM_USE_ODOMETRY = getBool(section, "hectorslam_use_odometry")
	HECTORSLAM_USE_LIDAR_CORRECTION = getBool(section, "hectorslam_use_lidar_correction")

	MOTORS_COM_NAME = getString(section, "motors_com_name")
	MOTORS_BAUD_RATE = getInt(section, "motors_baud_rate")
	MOTORS_RANGE_MIN = getInt(section, "motors_range_min")
	MOTORS_RANGE_MAX = getInt(section, "motors_range_max")

	MAP_STORAGE_ROOT = ASSETS_ROOT + getString(section, "map_storage_root")

	COLLISION_DETECTION_ANGLE = getFloat64(section, "collision_detection_angle")
	COLLISION_DETECTION_RADIUS = getFloat64(section, "collision_detection_radius")

	LOOKAHEAD_DISTANCE = getFloat64(section, "lookahead_distance")
	LOOKAHEAD_P = getFloat64(section, "lookahead_p")
	LOOKAHEAD_U = getFloat64(section, "lookahead_u")
	LOOKAHEAD_TDELTA = getInt(section, "lookahead_tdelta")

	ASTAR_MAX_ITERATIONS = getInt(section, "astar_max_iterations")
	ASTAR_UNKNOWN_PUNISH = getFloat64(section, "astar_unknown_punish")
	ASTAR_SMOOTHING_DATA_WEIGHT = getFloat64(section, "astar_smoothing_data_weight")
	ASTAR_SMOOTHING_SMOOTH_WEIGHT = getFloat64(section, "astar_smoothing_smooth_weight")
	ASTAR_CHECK_RADIUS = getFloat64(section, "astar_check_radius")
	ASTAR_SHRINK_FACTOR = getInt(section, "astar_shrink_factor")
}

func getString(section, option string) string {
	value, err := ConfigFile.GetString(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

func getInt(section, option string) int {
	value, err := ConfigFile.GetInt(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

func getBool(section, option string) bool {
	value, err := ConfigFile.GetBool(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

func getFloat64(section, option string) float64 {
	value, err := ConfigFile.GetFloat64(section, option)
	if err != nil {
		panic(err)
	}
	return value
}

// General
var (
	ASSETS_ROOT string
)

// Web config
var (
	API_URL         string
	STATIC_URL      string
	STATIC_ROOT     string
	TEMPLATE_ROOT   string
	BASE_TEMPLATE   string
	WEB_ADDRESS     string
	REALM           string
	USERNAME        string
	PASSPHRASE      string
	LOG_UPDATE_RATE int
)

// Datasets
var (
	DATASETS_ROOT string
)

// Sensor
var (
	SENSORLOGS_ROOT        string
	SENSORLOGS_TIME_FORMAT string
	USE_LIDAR              bool
	LIDAR_COM_NAME         string
	LIDAR_BAUD_RATE        int
	LIDAR_NUM_DISTANCES    int
	LIDAR_RADIAL_SPAN      float64
	LIDAR_MAX_DISTANCE     float64
	LIDAR_POSITION_X       float64
	LIDAR_POSITION_Y       float64
)

// Driving
var (
	MAX_SPEED float64
)

// Robot
var (
	ROBOT_BASE_WIDTH   float64
	ROBOT_WHEEL_RADIUS float64
	ROBOT_WHEEL_RATIO  float64
	ROBOT_ODOMETRY_PPR int
)

// Slam algorithm
var (
	SLAM_ALGORITHM string
)

// TinySLAM
var (
	TINYSLAM_GRIDMAP_SIZE          int
	TINYSLAM_GRIDMAP_RESOLUTION    int
	TINYSLAM_SIGMA_XY              float64
	TINYSLAM_SIGMA_THETA           float64
	TINYSLAM_HOLE_WIDTH            int
	TINYSLAM_MONTECARLO_ITERATIONS int
)

// HectorSLAM
var (
	HECTORSLAM_GRIDMAP_SIZE_X            int
	HECTORSLAM_GRIDMAP_SIZE_Y            int
	HECTORSLAM_GRIDMAP_RESOLUTION        float64
	HECTORSLAM_GRIDMAP_START_X           float64
	HECTORSLAM_GRIDMAP_START_Y           float64
	HECTORSLAM_LEVELS                    int
	HECTORSLAM_UPDATE_FACTOR_FREE        float64
	HECTORSLAM_UPDATE_FACTOR_OCCUPIED    float64
	HECTORSLAM_MAP_UPDATE_MIN_ANGLE_DIFF float64
	HECTORSLAM_MAP_UPDATE_MIN_DIST_DIFF  float64
	HECTORSLAM_USE_ODOMETRY              bool
	HECTORSLAM_USE_LIDAR_CORRECTION      bool
)

// Motors
var (
	MOTORS_COM_NAME  string
	MOTORS_BAUD_RATE int
	MOTORS_RANGE_MIN int
	MOTORS_RANGE_MAX int
)

// Map storage
var (
	MAP_STORAGE_ROOT string
)

// Collision detection
var (
	COLLISION_DETECTION_RADIUS float64
	COLLISION_DETECTION_ANGLE  float64
)

// Odometry
var (
	USE_ODOMETRY       bool
	ODOMETRY_COM_NAME  string
	ODOMETRY_BAUD_RATE int
)

// Lookahead
var (
	LOOKAHEAD_DISTANCE float64
	LOOKAHEAD_P        float64
	LOOKAHEAD_U        float64
	LOOKAHEAD_TDELTA   int
)

// Astar pathplanning
var (
	ASTAR_MAX_ITERATIONS          int
	ASTAR_UNKNOWN_PUNISH          float64
	ASTAR_SMOOTHING_DATA_WEIGHT   float64
	ASTAR_SMOOTHING_SMOOTH_WEIGHT float64
	ASTAR_CHECK_RADIUS            float64
	ASTAR_SHRINK_FACTOR           int
)
