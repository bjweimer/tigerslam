package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"robot/config"
	"time"

	"github.com/tarm/serial"
)

var lidardist = make([]int64, 360) // I think this is correct..............

// Request types for the URG
type RequestType int32

const (
	URG_GD = iota
	URG_GD_INTENSITY
	URG_GS
	URG_MD
	URG_MD_INTENSITY
	URG_MS
)

//const (
//	INFINITY_TIMES = C.UrgInfinityTimes
//)

// Urg holds the connection information and the C urg structure, and its methods
// implement the communication with an URG device.
type Urg struct {
	// The device port where the URG device is, e.g. "COM7"
	device string
	// The baud rate of the port, e.g. 115200
	baudRate int
	// The minimum step to measure. See manual for more explaination
	minStep int
	// The maximum measured step
	maxStep int
	// The minimum measureable distance
	minDistance int
	// The maximum measureable distance
	maxDistance int

	//dataMax int
	//	curg    *C.urg_t
	//cdata *C.long
}

//  read_int64 converts little endian bytes to an int64
func read_int64(data []byte) (ret int64) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

//  readLidar opens communication with the Neato lidar and then continuously reads data into lidardist[]
func readLidar() {
	//fmt.Printf("Lidar COM = %v\n", config.LIDAR_COM_NAME)
	//c := &serial.Config{Name: "COM4", Baud: 115200}
	c := &serial.Config{Name: config.LIDAR_COM_NAME, Baud: 115200}
	s, _ := serial.OpenPort(c)
	buf := make([]byte, 1980)

	// start the lidar data sampling loop to run forever
	for {
		n, err := s.Read(buf) // read the lidar serial buffer, this isn't a constant amount of data
		if err != nil {
			log.Fatal(err)
		}
		for j := 0; j < n; j++ {
			if buf[j] == 0xFA { // is it a start byte
				j++

				if j >= 1980 { // keep the index within bounds
					continue
				}

				if buf[j] >= 0xA0 && buf[j] <= 0xF9 { // is the next byte a packet number
					packet := 0
					packet = int(buf[j] - 0xA0) // subtract 0xA0 to get the actual packet number

					angle0 := packet * 4 // which 4 angles are in this packet
					angle1 := angle0 + 1
					angle2 := angle0 + 2
					angle3 := angle0 + 3

					if j+16 >= 1980 { // keep the index within bounds
						continue
					}

					if (buf[j+4] & 0x80) != 0 { // check to see if the "invalid data" flag is set
						continue
					}

					//  unpack the data in little endian format from two consecutive data bytes
					d0l := buf[j+3]
					d0h := buf[j+4] & 0x3F // only the lower 6 bits are data
					dist0 := read_int64([]byte{d0l, d0h, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

					d1l := buf[j+7]
					d1h := buf[j+8] & 0x3F
					dist1 := read_int64([]byte{d1l, d1h, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

					d2l := buf[j+11]
					d2h := buf[j+12] & 0x3F
					dist2 := read_int64([]byte{d2l, d2h, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

					d3l := buf[j+15]
					d3h := buf[j+16] & 0x3F
					dist3 := read_int64([]byte{d3l, d3h, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

					//  if the data passes this screening test, put it in the lidardist array
					if dist0 >= 150 {
						lidardist[angle0] = dist0
					}
					if dist1 >= 150 {
						lidardist[angle1] = dist1
					}
					if dist2 >= 150 {
						lidardist[angle2] = dist2
					}
					if dist3 >= 150 {
						lidardist[angle3] = dist3
					}
				}
			}
		}
	}
}

// Make URG from device string and baud rate
func MakeUrg(device string, baudRate, minStep, maxStep int) *Urg {
	u := &Urg{
		device:   device,
		baudRate: baudRate,
		minStep:  minStep,
		maxStep:  maxStep,
		//		curg:     new(C.urg_t),
	}

	return u
}

// Make the default URG from config
func MakeDefaultUrg() *Urg {
	return MakeUrg(config.LIDAR_COM_NAME, config.LIDAR_BAUD_RATE, 0, 360) // was 44, 725
}

// Connect to the URG
func (u *Urg) Connect() error {
	// Debug: fmt.Println("Connect function") this runs once when the lidar is connected
	var err error

	go readLidar() //  start the lidar reading thread
	// Debug: fmt.Println("readLIDAR thread started") this goroutine works

	// Find minimum measureable distance
	u.minDistance, err = u.MinDistance()
	if err != nil {
		return err
	}

	// Find maximum measureable distance
	u.maxDistance, err = u.MaxDistance()
	if err != nil {
		return err
	}

	// Find DataMax
	//u.dataMax, err = u.DataMax()
	//if err != nil {
	//	return err
	//}

	// Set up cdata, the array the c writes distance data to
	//var dummy C.long
	//size := int(unsafe.Sizeof(dummy))
	//u.cdata = (*C.long)(C.malloc(C.size_t(size * u.dataMax)))

	return nil
}

// Disconnect from the URG
func (u *Urg) Disconnect() {
	fmt.Println("Disconnect function")
	//	C.urg_disconnect(u.curg)
}

// Check if an URG is connected
func (u *Urg) IsConnected() bool {
	fmt.Println("IsConnected function")
	//	return int(C.urg_isConnected(u.curg)) == 1
	return true
}

// Get the URG Model Type as a string
func (u *Urg) GetModel() string {
	fmt.Println("GetModel function")
	//	return C.GoString(C.urg_model(u.curg))
	return "Neato LIDAR"
}

// Get Error
func (u *Urg) Error() error {
	fmt.Println("Error function")
	//	return errors.New(C.GoString(C.urg_error(u.curg)))
	return nil
}

// Set Skip Lines, a.k.a. Clustering
//
// The volume of aquired data can be reduced by skipping lines. This is the same
// as "clustering" in the URG product manual. Skipping lines will cause the URG
// to return the minimum value over the cluster of values, thus reducing the
// number of data points returned by a factor.
func (u *Urg) SetSkipLines(lines int) error {
	fmt.Println("SetSkipLines function")
	/*	if int(C.urg_setSkipLines(u.curg, C.int(lines))) < 0 {
			return u.Error()
		}
	*/
	return nil
}

// Get data max
func (u *Urg) DataMax() (int, error) {
	fmt.Println("DataMax function")
	/*	dataMax := int(C.urg_dataMax(u.curg))

		if dataMax < 0 {
			return dataMax, u.Error()
		}

		return dataMax, nil
	*/
	max := 360
	return max, nil
}

// Get minimum measureable distance
func (u *Urg) MinDistance() (int, error) {
	// Debug: fmt.Println("MinDistance function") this gets called once when the lidar is connected
	//	minDistance := int(C.urg_minDistance(u.curg))
	minDistance := 140

	if minDistance < 0 {
		return minDistance, u.Error()
	}

	return minDistance, nil
}

// Get maximum measureable distance
func (u *Urg) MaxDistance() (int, error) {
	// Debug: fmt.Println("MaxDistance function") this gets called once when the lidar is connected
	//	maxDistance := int(C.urg_maxDistance(u.curg))
	maxDistance := 4000

	if maxDistance < 0 {
		return maxDistance, u.Error()
	}

	return maxDistance, nil
}

// Aquire the latest timestamp
func (u *Urg) RecentTimestamp() (int, error) { // I made this up**************************
	fmt.Println("RecentTimestamp function")
	t := time.Now()
	timestamp := t.Nanosecond()
	return timestamp, nil
}

/*	func (u *Urg) RecentTimestamp() (int, error) {
		/*	timestamp := int(C.urg_recentTimestamp(u.curg))

			if timestamp < 0 {
				return timestamp, u.Error()
			}

			return timestamp, nil
}
*/

// Request Infinite Data
//
// Use the URG MD word and request INFINITY_TIMES data. Subsequent calls to
// ReceiveData() will return fresh data.
func (u *Urg) RequestInfiniteData() error {
	// Debug: fmt.Println("RequestInfiniteData function") this gets called once when the lidar is started running
	// Set capture times
	/*	if C.urg_setCaptureTimes(u.curg, INFINITY_TIMES) < 0 {
			return u.Error()
		}

		// Request data
		if C.urg_requestData(u.curg, URG_MD, C.int(u.minStep), C.int(u.maxStep)) < 0 {
			return u.Error()
		}
	*/
	return nil
}

// Read lidar data
func (u *Urg) ReceiveData() ([]float64, error) {
	data := make([]float64, 241) //  data = 240 degrees

	for i := 0; i < 210; i++ {
		data[i] = float64(lidardist[i+150])
	}

	for i := 210; i < 240; i++ {
		data[i] = float64(lidardist[i-210])
	}
	// Needed to put in a time delay here to prevent channel blocking.......
	time.Sleep(time.Millisecond * 100)
	return data, nil
}
