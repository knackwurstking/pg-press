package shared

import (
	"fmt"
	"time"
)

const (
	DateFormat = "02.01.2006"
	TimeFormat = "15:04"
)

type UnixMilli int64

func NewUnixMilli(t time.Time) UnixMilli {
	return UnixMilli(t.UnixMilli())
}

func (um UnixMilli) FormatDate() string {
	if um == 0 {
		return ""
	}
	return time.UnixMilli(int64(um)).Format(DateFormat)
}

func (um UnixMilli) FormatDateTime() string {
	if um == 0 {
		return ""
	}
	return time.UnixMilli(int64(um)).Format(fmt.Sprintf("%s %s", DateFormat, TimeFormat))
}
