package gzutil

import (
	"fmt"
	"math"
)

// CalculateDistance 返回两点间距离，单位米
func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	radius := 6371000.0 // 地球半径，单位米
	dLat := DegToRad(lat2 - lat1)
	dLng := DegToRad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(DegToRad(lat1))*math.Cos(DegToRad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := radius * c

	return distance
}

// DegToRad 角度转弧度
func DegToRad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func FormatDistance(distance float64) string {
	if distance < 1000 {
		return fmt.Sprintf("%.0f", distance)
	}
	return fmt.Sprintf("%.2f", distance/1000)
}

// FormatDistance 格式化距离显示，显示为m或km
func FormatDistanceWithEn(distance float64) string {
	if distance < 1000 {
		return fmt.Sprintf("%.0f m", distance)
	}
	return fmt.Sprintf("%.2f km", distance/1000)
}

func FormatDistanceWithZh(distance float64) string {
	if distance < 1000 {
		return fmt.Sprintf("%.0f 米", distance)
	}
	return fmt.Sprintf("%.2f 公里", distance/1000)
}
