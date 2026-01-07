package gztime

import "time"

func TimePtr(t time.Time) *time.Time {
	return &t
}

func TimeInTimeRange(checkTime, start, end time.Time) bool {
	current := time.Date(0, 1, 1,
		checkTime.Hour(), checkTime.Minute(), checkTime.Minute(), 0, time.Local)

	s := time.Date(0, 1, 1, start.Hour(), start.Minute(), start.Second(), 0, time.Local)
	e := time.Date(0, 1, 1, end.Hour(), end.Minute(), end.Second(), 0, time.Local)

	return current.After(s) && current.Before(e)
}

// 获取某一天的开始和结束时间
func GetDayRange(t time.Time) (time.Time, time.Time) {
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	endTime := startTime.AddDate(0, 0, 1)

	return startTime, endTime
}
