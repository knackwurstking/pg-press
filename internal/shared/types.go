package shared

import (
	"fmt"
	"time"
)

// EditorType represents the type of editor for different entities
type EditorType string

// PressType represents the type of press machine
type PressType string

// PressNumber represents a press machine number
type PressNumber int8

// Slot represents a slot type (for tool cassette slots or press positions)
type Slot int

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

// Constants for slot types
const (
	SlotUnknown           Slot = 0
	SlotPressUp           Slot = 1
	SlotPressDown         Slot = 2
	SlotUpperToolCassette Slot = 10
	//SlotLowerToolCassette Slot = 20
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
	switch p {
	case 0, 2, 3, 4, 5:
		return true
	default:
		return false
	}
}

// Slot methods
func (p Slot) String() string {
	switch p {
	case SlotPressUp:
		return "UP"
	case SlotPressDown:
		return "DOWN"
	default:
		return "UNKNOWN"
	}
}

func (p Slot) German() string {
	switch p {
	case SlotPressUp:
		return "Oben"
	case SlotPressDown:
		return "Unten"
	default:
		return "?"
	}
}

func (um UnixMilli) FormatDate() string {
	return time.UnixMilli(int64(um)).Format(DateFormat)
}

func (um UnixMilli) FormatDateTime() string {
	return time.UnixMilli(int64(um)).Format(fmt.Sprintf("%s %s", DateFormat, TimeFormat))
}
