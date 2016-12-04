package datacontainer

import ()

type DataContainer struct {
	dataPoints [][2]float64
	origo      [2]float64
}

func MakeDataContainer(size int) *DataContainer {
	dc := new(DataContainer)
	dc.dataPoints = make([][2]float64, size)
	dc.origo = [2]float64{0, 0}
	return dc
}

func (dc *DataContainer) SetFrom(other *DataContainer, factor float64) {
	dc.origo = other.GetOrigo()
	dc.origo[0] *= factor
	dc.origo[1] *= factor

	if len(dc.dataPoints) != len(other.dataPoints) {
		dc.dataPoints = make([][2]float64, len(other.dataPoints))
	}

	for i := range other.dataPoints {
		dc.dataPoints[i][0] = other.dataPoints[i][0] * factor
		dc.dataPoints[i][1] = other.dataPoints[i][1] * factor
	}
}

func (dc *DataContainer) Add(dataPoint [2]float64) {
	dc.dataPoints = append(dc.dataPoints, dataPoint)
}

func (dc *DataContainer) Clear() {
	dc.dataPoints = make([][2]float64, 0)
}

func (dc *DataContainer) GetSize() int {
	return len(dc.dataPoints)
}

func (dc *DataContainer) GetVecEntry(index int) [2]float64 {
	return dc.dataPoints[index]
}

func (dc *DataContainer) GetOrigo() [2]float64 {
	return dc.origo
}

func (dc *DataContainer) SetOrigo(origoIn [2]float64) {
	dc.origo = origoIn
}
