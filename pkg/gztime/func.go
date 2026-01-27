package gztime

import (
	"fmt"
	"time"
)

var layouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"2006-01",
	"2006",
}

func ParseFlexibleTime(s string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.Local
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", s)
}

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

func DayRange(t time.Time) (time.Time, time.Time) {
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	endTime := startTime.AddDate(0, 0, 1)

	return startTime, endTime
}

func WeekRange(t time.Time) (time.Time, time.Time) {
	loc := t.Location()
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	monday := time.Date(t.Year(), t.Month(), t.Day()-(weekday-1), 0, 0, 0, 0, loc)
	sunday := monday.AddDate(0, 0, 7).Add(-time.Second)

	return monday, sunday
}

func MonthRange(month string, loc *time.Location) (time.Time, time.Time, error) {
	t, err := ParseFlexibleTime(month, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	nextMonth := start.AddDate(0, 1, 0)
	end := nextMonth.Add(-time.Nanosecond)

	return start, end, nil
}

func YearRange(year string, loc *time.Location) (time.Time, time.Time, error) {
	t, err := ParseFlexibleTime(year, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	start := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
	nextYear := start.AddDate(1, 0, 0)
	end := nextYear.Add(-time.Nanosecond)

	return start, end, nil
}
