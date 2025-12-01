package gzdb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

// --------------------------------------------------------------------------------------------------------------------
//
//	一个通用的、可用于 GORM 的 JSON 数组类型，T 可以是 string, int, int64, int32, 或任何可 JSON 序列化的类型
//
// --------------------------------------------------------------------------------------------------------------------
type JSONList[T any] []T

func (j JSONList[T]) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return "[]", nil
	}

	return json.Marshal(j)
}

func (j *JSONList[T]) Scan(value interface{}) error {
	if value == nil {
		*j = []T{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for JSONList: %T", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*j = []T{}
		return nil
	}

	if err := json.Unmarshal(bytes, j); err != nil {
		return fmt.Errorf("failed to unmarshal JSONList: %w", err)
	}

	return nil
}

func (j JSONList[T]) ToSlice() []T {
	if j == nil {
		return []T{}
	}

	return []T(j)
}

// --------------------------------------------------------------------------------------------------------------------
//
//	素材存储时自动移除、查询时自动增加oss域名信息 Start
//
// --------------------------------------------------------------------------------------------------------------------
type MaterialURL string

func (u MaterialURL) Value() (driver.Value, error) {
	s := string(u)
	if s == "" {
		return "", nil
	}

	return gzutil.RemoveDomain(s), nil
}

func (u *MaterialURL) Scan(value interface{}) error {
	if value == nil {
		*u = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			*u = ""
			return nil
		}
		*u = MaterialURL(gzutil.JoinDomain(gooze.Config.Oss.Url, v))
	case []byte:
		if len(v) == 0 {
			*u = ""
			return nil
		}
		*u = MaterialURL(gzutil.JoinDomain(gooze.Config.Oss.Url, string(v)))
	default:
		*u = ""
	}

	return nil
}

func (u MaterialURL) String() string {
	if u == "" {
		return ""
	}
	return string(u)
}

// --------------------------------------------------------------------------------------------------------------------
//
//	多素材存储时自动移除、查询时自动增加oss域名信息 Start
//
// --------------------------------------------------------------------------------------------------------------------
type MaterialList []string

func (i MaterialList) Value() (driver.Value, error) {
	if len(i) == 0 {
		return "[]", nil
	}

	cleaned := make([]string, 0, len(i))
	for _, v := range i {
		cleaned = append(cleaned, gzutil.RemoveDomain(v))
	}

	return json.Marshal(cleaned)
}

func (i *MaterialList) Scan(value interface{}) error {
	if value == nil {
		*i = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for MaterialList: %T", value)
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*i = []string{}
		return nil
	}

	if bytes[0] == '[' {
		var relativePaths []string
		if err := json.Unmarshal(bytes, &relativePaths); err != nil {
			return fmt.Errorf("failed to unmarshal MaterialList JSON array: %w", err)
		}

		finalPaths := make([]string, 0, len(relativePaths))
		for _, p := range relativePaths {
			finalPaths = append(finalPaths, gzutil.JoinDomain(gooze.Config.Oss.Url, p))
		}
		*i = finalPaths
	} else {
		*i = []string{string(bytes)}
	}

	return nil
}

func (l MaterialList) ToSlice() []string {
	return []string(l)
}

// --------------------------------------------------------------------------------------------------------------------
//
//	处理 MySQL 为 time 类型的，自动转为 Go 的 time.Time类型
//
// --------------------------------------------------------------------------------------------------------------------
type NullableTime struct {
	Time     time.Time
	Valid    bool
	Location *time.Location
}

var DefaultLocation = time.Local

func (nt *NullableTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	var timeStr string
	switch v := value.(type) {
	case []byte:
		timeStr = string(v)
	case string:
		timeStr = v
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	if nt.Location == nil {
		nt.Location = DefaultLocation
	}
	today := time.Now().In(nt.Location).Format("2006-01-02")
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", today+" "+timeStr, nt.Location)
	if err != nil {
		return err
	}
	nt.Time = parsed
	nt.Valid = true
	return nil
}

func (nt NullableTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time.Format("15:04:05"), nil
}

func (nt *NullableTime) SetLocation(loc *time.Location) *NullableTime {
	nt.Location = loc
	if nt.Valid {
		nt.Time = time.Date(0, 1, 1, nt.Time.Hour(), nt.Time.Minute(), nt.Time.Second(), 0, loc)
	}
	return nt
}

func NewNullableTimeFromTime(t time.Time, loc *time.Location) NullableTime {
	if loc == nil {
		loc = DefaultLocation
	}
	return NullableTime{
		Time:     time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, loc),
		Valid:    true,
		Location: loc,
	}
}

func NewNullableTimeFromString(timeStr string, loc *time.Location) (NullableTime, error) {
	if loc == nil {
		loc = DefaultLocation
	}
	t, err := time.ParseInLocation("15:04:05", timeStr, loc)
	if err != nil {
		return NullableTime{}, err
	}
	return NullableTime{
		Time:     time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, loc),
		Valid:    true,
		Location: loc,
	}, nil
}

func (nt NullableTime) ToTime() time.Time {
	if !nt.Valid {
		return time.Time{}
	}
	return nt.Time
}
