package occbase

import (
	"math"

	"github.com/skelterjohn/go.matrix"

	"hectormapping/datacontainer"
	"hectormapping/map/cache"
	"hectormapping/map/gridmap"
	"hectormapping/utils"
)

type OccGridMapUtil struct {
	intensities     [4]float64
	cacheMethod     cache.CacheMethod
	concreteGridMap gridmap.OccGridMap
	samplePoints    [][3]float64
	//	size int
	mapObstacleThreshold float64
}

func MakeOccGridMapUtil(gridMap gridmap.OccGridMap, cacheMethod cache.CacheMethod) *OccGridMapUtil {
	ogmu := &OccGridMapUtil{
		concreteGridMap: gridMap,
		//		size: 0,
		mapObstacleThreshold: gridMap.GetObstacleThreshold(),
		cacheMethod:          cacheMethod,
	}

	ogmu.cacheMethod.SetMapSize(gridMap.GetMapDimensions())

	return ogmu
}

func (ogmu *OccGridMapUtil) GetWorldCoordsPose(mapPose [3]float64) [3]float64 {
	return ogmu.concreteGridMap.GetWorldCoordsPose(mapPose)
}

func (ogmu *OccGridMapUtil) GetMapCoordsPose(worldPose [3]float64) [3]float64 {
	return ogmu.concreteGridMap.GetMapCoordsPose(worldPose)
}

func (ogmu *OccGridMapUtil) GetWorldCoordsPoint(mapPoint [2]float64) [2]float64 {
	return ogmu.concreteGridMap.GetWorldCoords(mapPoint)
}

func (ogmu *OccGridMapUtil) GetCompleteHessianDerivs(pose *[3]float64,
	dataPoints *datacontainer.DataContainer, H *matrix.DenseMatrix, dTr *matrix.DenseMatrix) {

	size := dataPoints.GetSize()

	transform := ogmu.GetTransformForState(*pose)

	sinRot := math.Sin(pose[2])
	cosRot := math.Cos(pose[2])

	*H = *matrix.Zeros(3, 3)
	*dTr = *matrix.Zeros(3, 1)

	for i := 0; i < size; i++ {

		currPoint := dataPoints.GetVecEntry(i)

		tmpMatrix, err := transform.TimesDense(matrix.MakeDenseMatrix(append(currPoint[:], 1), 3, 1))
		if err != nil {
			panic(err)
		}
		transformedCurrPoint := [2]float64{tmpMatrix.Get(0, 0), tmpMatrix.Get(1, 0)}

		transformedPointData := ogmu.InterpMapValueWithDerivatives(transformedCurrPoint)

		funVal := 1.0 - transformedPointData[0]

		dTr.Set(0, 0, dTr.Get(0, 0)+transformedPointData[1]*funVal)
		dTr.Set(1, 0, dTr.Get(1, 0)+transformedPointData[2]*funVal)

		rotDeriv := ((-sinRot*currPoint[0]-cosRot*currPoint[1])*transformedPointData[1] + (cosRot*currPoint[0]-sinRot*currPoint[1])*transformedPointData[2])

		dTr.Set(2, 0, dTr.Get(2, 0)+rotDeriv*funVal)

		H.Set(0, 0, H.Get(0, 0)+transformedPointData[1]*transformedPointData[1])
		H.Set(1, 1, H.Get(1, 1)+transformedPointData[2]*transformedPointData[2])
		H.Set(2, 2, H.Get(2, 2)+rotDeriv*rotDeriv)

		H.Set(0, 1, H.Get(0, 1)+transformedPointData[1]*transformedPointData[2])
		H.Set(0, 2, H.Get(0, 2)+transformedPointData[1]*rotDeriv)
		H.Set(1, 2, H.Get(1, 2)+transformedPointData[2]*rotDeriv)
	}

	H.Set(1, 0, H.Get(0, 1))
	H.Set(2, 0, H.Get(0, 2))
	H.Set(2, 1, H.Get(1, 2))
}

func (ogmu *OccGridMapUtil) GetCovarianceForPose(mapPose [3]float64, dataPoints *datacontainer.DataContainer) *matrix.DenseMatrix {
	deltaTransX := 1.5
	deltaTransY := 1.5
	deltaAng := 0.05

	x := mapPose[0]
	y := mapPose[1]
	ang := mapPose[2]

	sigmaPoints := matrix.Zeros(3, 7)

	sigmaPoints.Set(0, 0, x+deltaTransX)
	sigmaPoints.Set(1, 0, y)
	sigmaPoints.Set(2, 0, ang)

	sigmaPoints.Set(0, 1, x-deltaTransX)
	sigmaPoints.Set(1, 1, y)
	sigmaPoints.Set(2, 1, ang)

	sigmaPoints.Set(0, 2, x)
	sigmaPoints.Set(1, 2, y+deltaTransY)
	sigmaPoints.Set(2, 2, ang)

	sigmaPoints.Set(0, 3, x)
	sigmaPoints.Set(1, 3, y-deltaTransY)
	sigmaPoints.Set(2, 3, ang)

	sigmaPoints.Set(0, 4, x)
	sigmaPoints.Set(1, 4, y)
	sigmaPoints.Set(2, 4, ang+deltaAng)

	sigmaPoints.Set(0, 5, x)
	sigmaPoints.Set(1, 5, y)
	sigmaPoints.Set(2, 5, ang-deltaAng)

	sigmaPoints.Set(0, 6, mapPose[0])
	sigmaPoints.Set(1, 6, mapPose[1])
	sigmaPoints.Set(2, 6, mapPose[2])

	likelihoods := matrix.Zeros(7, 1)

	likelihoods.Set(0, 0, ogmu.GetLikelihoodForState([3]float64{x + deltaTransX, y, ang}, dataPoints))
	likelihoods.Set(1, 0, ogmu.GetLikelihoodForState([3]float64{x - deltaTransX, y, ang}, dataPoints))
	likelihoods.Set(2, 0, ogmu.GetLikelihoodForState([3]float64{x, y + deltaTransY, ang}, dataPoints))
	likelihoods.Set(3, 0, ogmu.GetLikelihoodForState([3]float64{x, y - deltaTransY, ang}, dataPoints))
	likelihoods.Set(4, 0, ogmu.GetLikelihoodForState([3]float64{x, y, ang + deltaAng}, dataPoints))
	likelihoods.Set(5, 0, ogmu.GetLikelihoodForState([3]float64{x, y, ang - deltaAng}, dataPoints))
	likelihoods.Set(6, 0, ogmu.GetLikelihoodForState([3]float64{x, y, ang}, dataPoints))

	likelihoodsArray := likelihoods.Array()
	likelihoodsSum := 0.0
	for i := range likelihoodsArray {
		likelihoodsSum += likelihoodsArray[i]
	}

	invLhNormalizer := 1.0 / likelihoodsSum

	mean := matrix.Zeros(3, 1)

	for i := 0; i < 7; i++ {
		sigmaCol := sigmaPoints.GetColVector(i)
		sigmaCol.Scale(likelihoods.Get(i, 0))
		err := mean.AddDense(sigmaCol)
		if err != nil {
			panic(err)
		}
	}

	mean.Scale(invLhNormalizer)

	covMatrixMap := matrix.Zeros(3, 3)

	for i := 0; i < 7; i++ {
		sigPointMinusMean := sigmaPoints.GetColVector(i)
		sigPointMinusMean.Minus(mean)

		add, err := sigPointMinusMean.TimesDense(sigPointMinusMean.Transpose())
		if err != nil {
			panic(err)
		}
		add.Scale(likelihoods.Get(i, 0) * invLhNormalizer)

		covMatrixMap.AddDense(add)
	}

	return covMatrixMap
}

func (ogmu *OccGridMapUtil) GetCovMatrixWorldCoords(covMatMap *matrix.DenseMatrix) *matrix.DenseMatrix {
	covMatWorld := matrix.Zeros(3, 3)

	scaleTrans := ogmu.concreteGridMap.GetCellLength()
	scaleTransSq := scaleTrans * scaleTrans

	covMatWorld.Set(0, 0, covMatMap.Get(0, 0)*scaleTransSq)
	covMatWorld.Set(1, 1, covMatMap.Get(1, 1)*scaleTransSq)

	covMatWorld.Set(1, 0, covMatMap.Get(1, 0)*scaleTransSq)
	covMatWorld.Set(0, 1, covMatWorld.Get(1, 0))

	covMatWorld.Set(2, 0, covMatMap.Get(2, 0)*scaleTrans)
	covMatWorld.Set(0, 2, covMatWorld.Get(2, 0))

	covMatWorld.Set(2, 1, covMatMap.Get(2, 1)*scaleTrans)
	covMatWorld.Set(1, 2, covMatWorld.Get(2, 1))

	covMatWorld.Set(2, 2, covMatMap.Get(2, 2))

	return covMatWorld
}

func (ogmu *OccGridMapUtil) GetLikelihoodForState(state [3]float64, dataPoints *datacontainer.DataContainer) float64 {
	resid := ogmu.GetResidualForState(state, dataPoints)
	return ogmu.GetLikelihoodForResidual(resid, dataPoints.GetSize())
}

func (ogmu *OccGridMapUtil) GetLikelihoodForResidual(residual float64, numDataPoints int) float64 {
	//	sizef := float64(numDataPoints)

	numDataPointsA := float64(numDataPoints)
	sizef := float64(numDataPointsA)

	return 1.0 - (residual / sizef)
}

func (ogmu *OccGridMapUtil) GetResidualForState(state [3]float64, dataPoints *datacontainer.DataContainer) float64 {
	size := dataPoints.GetSize()

	stepSize := 1
	residual := 0.0

	transform := ogmu.GetTransformForState(state)

	for i := 0; i < size; i += stepSize {
		vecEntry := dataPoints.GetVecEntry(i)
		tempMatrix, err := transform.TimesDense(matrix.MakeDenseMatrix(append(vecEntry[:], 1), 3, 1))
		if err != nil {
			panic(err)
		}
		transformedVecEntry := [2]float64{tempMatrix.Get(0, 0), tempMatrix.Get(1, 0)}

		funval := 1.0 - ogmu.InterpMapValue(transformedVecEntry)
		residual += funval
	}

	return residual
}

// Get interpolated map value. Map values are interpolated using a bilinear
// filtering scheme. They are used estimating occupancy probabilities and
// derivatives. Intuitively, the grid map cell values can be viewed as samples
// of an underlying continuous probability distribution.
func (ogmu *OccGridMapUtil) InterpMapValue(coords [2]float64) float64 {

	// Check if coords are within map limits
	if ogmu.concreteGridMap.PointOutOfMapBounds(coords) {
		//		logger.Println("Tried to check interpolated value for out-of-bounds point")
		return 0
	}

	// Map coords are always positive, floor them by casting to int
	indMin := [2]int{int(coords[0]), int(coords[1])}

	// Get factors for bilinear interpolation
	factors := [2]float64{coords[0] - float64(indMin[0]), coords[1] - float64(indMin[1])}

	sizeX := ogmu.concreteGridMap.GetSizeX()

	index := indMin[1]*sizeX + indMin[0]

	// Get grid values for the 4 grid points surrounding the current coords.
	// Check cached data first, if not contained filter gridPoint with gaussian
	// and store in cache.
	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[0]) {
		ogmu.intensities[0] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[0])
	}

	index++

	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[1]) {
		ogmu.intensities[1] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[1])
	}

	index += sizeX - 1

	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[2]) {
		ogmu.intensities[2] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[2])
	}

	index++

	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[3]) {
		ogmu.intensities[3] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[3])
	}

	xFacInv := 1.0 - factors[0]
	yFacInv := 1.0 - factors[1]

	return ((ogmu.intensities[0]*xFacInv + ogmu.intensities[1]*factors[0]) * yFacInv) +
		((ogmu.intensities[2]*xFacInv + ogmu.intensities[3]*factors[0]) * factors[1])
}

func (ogmu *OccGridMapUtil) InterpMapValueWithDerivatives(coords [2]float64) [3]float64 {

	// Check if coords are within map limits
	if ogmu.concreteGridMap.PointOutOfMapBounds(coords) {
		//		logger.Println("Tried to check interpolated value for out-of-bounds point")
		return [3]float64{0, 0, 0}
	}

	// Map coords are always positive, floor them by casting to int
	indMin := [2]int{int(coords[0]), int(coords[1])}

	// Get factors for bilinear interpolation
	factors := [2]float64{coords[0] - float64(indMin[0]), coords[1] - float64(indMin[1])}

	sizeX := ogmu.concreteGridMap.GetSizeX()

	index := indMin[1]*sizeX + indMin[0]

	// Get grid values for the 4 grid points surrounding the current coords.
	// Check cached data first, if not contained filter gridPoint with gaussian
	// and store in cache.
	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[0]) {
		ogmu.intensities[0] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[0])
	}

	index++

	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[1]) {
		ogmu.intensities[1] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[1])
	}

	index += sizeX - 1

	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[2]) {
		ogmu.intensities[2] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[2])
	}

	index++

	if !ogmu.cacheMethod.ContainsCachedData(index, &ogmu.intensities[3]) {
		ogmu.intensities[3] = ogmu.GetUnfilteredGridPointByIndex(index)
		ogmu.cacheMethod.CacheData(index, ogmu.intensities[3])
	}

	dx1 := ogmu.intensities[0] - ogmu.intensities[1]
	dx2 := ogmu.intensities[2] - ogmu.intensities[3]

	dy1 := ogmu.intensities[0] - ogmu.intensities[2]
	dy2 := ogmu.intensities[1] - ogmu.intensities[3]

	xFacInv := 1.0 - factors[0]
	yFacInv := 1.0 - factors[1]

	r1 := (ogmu.intensities[0]*xFacInv+ogmu.intensities[1]*factors[0])*yFacInv +
		(ogmu.intensities[2]*xFacInv+ogmu.intensities[3]*factors[0])*factors[1]
	r2 := -((dx1 * xFacInv) + (dx2 * factors[0]))
	r3 := -((dy1 * yFacInv) + (dy2 * factors[1]))

	return [3]float64{r1, r2, r3}
}

func (ogmu *OccGridMapUtil) GetUnfilteredGridPoint(gridCoords [2]int) float64 {
	return ogmu.concreteGridMap.GetGridProbabilityMap(gridCoords[0], gridCoords[1])
}

func (ogmu *OccGridMapUtil) GetUnfilteredGridPointByIndex(index int) float64 {
	return ogmu.concreteGridMap.GetGridProbabilityMapByIndex(index)
}

func (ogmu *OccGridMapUtil) GetTransformForState(transVector [3]float64) *matrix.DenseMatrix {
	return utils.TransformationMatrix2D(transVector)
}

func (ogmu *OccGridMapUtil) GetTranslationForState(transVector [3]float64) *matrix.DenseMatrix {
	point := [3]float64{transVector[0], transVector[1], 0}
	return utils.TransformationMatrix2D(point)
}

func (ogmu *OccGridMapUtil) ResetCachedData() {
	ogmu.cacheMethod.ResetCache()
}

func (ogmu *OccGridMapUtil) ResetSamplePoints() {
	ogmu.samplePoints = make([][3]float64, 0)
}

func (ogmu *OccGridMapUtil) GetSamplePoints() [][3]float64 {
	return ogmu.samplePoints
}
