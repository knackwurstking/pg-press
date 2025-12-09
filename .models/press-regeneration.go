package models

import (
	"time"

	"github.com/knackwurstking/pg-press/errors"
)

type PressRegenerationID int64

type PressRegeneration struct {
	ID          PressRegenerationID `json:"id"`
	PressNumber PressNumber         `json:"tool_id"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt time.Time           `json:"completed_at"`
	Reason      string              `json:"reason"`
}

func NewPressRegeneration(pn PressNumber, startedAt time.Time, reason string) *PressRegeneration {
	return &PressRegeneration{
		PressNumber: pn,
		StartedAt:   startedAt,
		Reason:      reason,
	}
}

func (r *PressRegeneration) Stop() {
	r.CompletedAt = time.Now()
}

func (r *PressRegeneration) StopAt(completedAt time.Time) {
	r.CompletedAt = completedAt
}

func (r *PressRegeneration) Validate() *errors.ValidationError {
	if !IsValidPressNumber(&r.PressNumber) {
		return errors.NewValidationError("invalid press_number: %d", r.PressNumber)
	}

	if r.StartedAt.IsZero() {
		return errors.NewValidationError("started_at cannot be zero")
	}

	if !r.CompletedAt.IsZero() && !r.CompletedAt.After(r.StartedAt) {
		return errors.NewValidationError("completed_at must be after started_at")
	}

	return nil
}
