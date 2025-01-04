package chart

import (
	"math"
	"time"
)

type Number interface {
	int | int64 | float32 | float64
}

// getRoundToForDelta 的作用是根据 delta 的大小返回一个适当的精度值，以便进行舍入或取整操作。
func getRoundToForDelta(delta float64) float64 {
	startingDeltaBound := math.Pow(10.0, 10.0)
	for cursor := startingDeltaBound; cursor > 0; cursor /= 10.0 {
		if delta > cursor {
			return cursor / 10.0
		}
	}
	return 0.0
}

// roundUp 将 value 向上舍入到最接近的 roundTo 的倍数
func roundUp(value, roundTo float64) float64 {
	d1 := math.Ceil(value / roundTo)
	return d1 * roundTo
}

// roundDown 将 value 向下舍入到最接近的 roundTo 的倍数
func roundDown(value, roundTo float64) float64 {
	d1 := math.Floor(value / roundTo)
	return d1 * roundTo
}

func abs[T Number](value T) T {
	if value < 0 {
		return -value
	}
	return value
}

func mean[T Number](values ...T) T {
	return sum(values...) / T(len(values))
}

func sum[T Number](values ...T) (total T) {
	for _, v := range values {
		total += v
	}
	return total
}

func degreesToRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180)
}

// rotateCoordinate 将二维坐标系中的点 (x,y) 以某个点 (cx,cy) 为旋转中心，围绕该中心旋转指定角度 θ（以弧度表示）。
// 旋转后的新坐标是 (rx,ry)，其中 (rx,ry) 是旋转后的坐标。
func rotateCoordinate(cx, cy, x, y int, thetaRadians float64) (rx, ry int) {
	tempX, tempY := float64(x-cx), float64(y-cy)
	rotatedX := tempX*math.Cos(thetaRadians) - tempY*math.Sin(thetaRadians)
	rotatedY := tempX*math.Sin(thetaRadians) + tempY*math.Cos(thetaRadians)
	rx = int(rotatedX) + cx
	ry = int(rotatedY) + cy
	return
}

// toFloat64 将时间对象转换为 float64 类型
func toFloat64(t time.Time) float64 {
	return float64(t.UnixNano())
}

// pointsToPixels 将字体大小（以点为单位）转换为像素值
func pointsToPixels(dpi, points float64) float64 {
	return (points * dpi) / 72.0
}
