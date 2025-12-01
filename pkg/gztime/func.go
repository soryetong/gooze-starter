package gztime

import "time"

func TimePtr(t time.Time) *time.Time {
	return &t
}

func GetTodayRange() (time.Time, time.Time) {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endTime := startTime.AddDate(0, 0, 1)

	return startTime, endTime
}

func TimeInTimeRange(checkTime, start, end time.Time) bool {
	current := time.Date(0, 1, 1,
		checkTime.Hour(), checkTime.Minute(), checkTime.Minute(), 0, time.Local)

	s := time.Date(0, 1, 1, start.Hour(), start.Minute(), start.Second(), 0, time.Local)
	e := time.Date(0, 1, 1, end.Hour(), end.Minute(), end.Second(), 0, time.Local)

	return current.After(s) && current.Before(e)
}
