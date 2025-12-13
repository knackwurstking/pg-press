package shared

import (
	"fmt"
	"slices"
	"time"
)

// EditorType represents the type of editor for different entities
type EditorType string

// PressType represents the type of press machine
type PressType string

// PressNumber represents a press machine number
type PressNumber int8

// EntityID represents a generic entity identifier
type EntityID int64

// TelegramID represents a Telegram user identifier
type TelegramID int64

// UnixMilli represents a Unix timestamp in milliseconds
type UnixMilli int64

// Constants for editor types
const (
	EditorTypeTroubleReport EditorType = "troublereport"
)

// Constants for press types
const (
	PressTypeSACMI PressType = "SACMI"
	PressTypeSITI  PressType = "SITI"
)

// Press number constants
const (
	PressNumber0 PressNumber = 0
	PressNumber2 PressNumber = 2
	PressNumber3 PressNumber = 3
	PressNumber4 PressNumber = 4
	PressNumber5 PressNumber = 5
)

var (
	AllPressesNumbers []PressNumber = []PressNumber{
		PressNumber0,
		PressNumber2,
		PressNumber3,
		PressNumber4,
		PressNumber5,
	}
)

// Constants for date/time formats
const (
	DateFormat = "02.01.2006"
	TimeFormat = "15:04"
)

func NewUnixMilli(t time.Time) UnixMilli {
	return UnixMilli(t.UnixMilli())
}

// EntityID methods
func (id EntityID) String() string {
	return fmt.Sprintf("%d", id)
}

// TelegramID methods
func (id TelegramID) String() string {
	return fmt.Sprintf("%d", id)
}

// PressNumber methods
func (p PressNumber) String() string {
	return fmt.Sprintf("%d", p)
}

func (p PressNumber) IsValid() bool {
	return slices.Contains(AllPressesNumbers, p)
}

func (um UnixMilli) FormatDate() string {
	return time.UnixMilli(int64(um)).Format(DateFormat)
}

func (um UnixMilli) FormatDateTime() string {
	return time.UnixMilli(int64(um)).Format(fmt.Sprintf("%s %s", DateFormat, TimeFormat))
}
