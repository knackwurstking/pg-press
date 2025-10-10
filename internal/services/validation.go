package services

import (
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// ValidateTool performs comprehensive tool validation
func ValidateTool(tool *models.Tool) error {
	if err := ValidateNotNil(tool, "tool"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(string(tool.Position), "position"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(tool.Type, "type"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(tool.Code, "code"); err != nil {
		return err
	}

	// Format validation would be handled by the model's own validation
	// since models.Format might not be a pointer type

	return nil
}

// ValidateModification performs comprehensive modification validation
func ValidateModification(mod *models.Modification[any]) error {
	if err := ValidateNotNil(mod, "modification"); err != nil {
		return err
	}

	if err := ValidateID(mod.UserID, "user"); err != nil {
		return err
	}

	if mod.Data == nil {
		return utils.NewValidationError("modification data cannot be nil")
	}

	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(limit, offset int) error {
	if limit < 0 {
		return utils.NewValidationError("limit cannot be negative")
	}

	if offset < 0 {
		return utils.NewValidationError("offset cannot be negative")
	}

	if limit > 1000 {
		return utils.NewValidationError("limit cannot exceed 1000")
	}

	return nil
}

// ValidatePressCycle performs comprehensive press cycle validation
func ValidatePressCycle(cycle *models.Cycle) error {
	if err := ValidateNotNil(cycle, "cycle"); err != nil {
		return err
	}

	if err := ValidatePressNumber(cycle.PressNumber); err != nil {
		return err
	}

	if err := ValidateID(cycle.ToolID, "tool"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(string(cycle.ToolPosition), "tool_position"); err != nil {
		return err
	}

	if err := ValidatePositive(cycle.TotalCycles, "total_cycles"); err != nil {
		return err
	}

	if cycle.Date.IsZero() {
		return utils.NewValidationError("date cannot be zero")
	}

	return nil
}

// ValidatePressNumber validates that a press number is within valid range (0-5)
func ValidatePressNumber(pressNumber models.PressNumber) error {
	if pressNumber < 0 || pressNumber > 5 {
		return utils.NewValidationError(fmt.Sprintf("press_number must be between 0 and 5, got: %d", pressNumber))
	}
	return nil
}

// ValidateToolRegeneration performs comprehensive tool regeneration validation
func ValidateToolRegeneration(regen *models.Regeneration) error {
	if err := ValidateNotNil(regen, "regeneration"); err != nil {
		return err
	}

	if err := ValidateID(regen.ToolID, "tool"); err != nil {
		return err
	}

	if err := ValidateID(regen.CycleID, "cycle"); err != nil {
		return err
	}

	// Reason is optional, but if provided should not be empty
	if regen.Reason != "" {
		if err := ValidateNotEmpty(regen.Reason, "reason"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateTroubleReport performs comprehensive trouble report validation
func ValidateTroubleReport(report *models.TroubleReport) error {
	if err := ValidateNotNil(report, "trouble_report"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(report.Title, "title"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(report.Content, "content"); err != nil {
		return err
	}

	// LinkedAttachments can be empty, but if present should contain valid IDs
	for i, attachmentID := range report.LinkedAttachments {
		if err := ValidateID(attachmentID, fmt.Sprintf("attachment_%d", i)); err != nil {
			return err
		}
	}

	return nil
}

// ValidateExistence checks if a value exists in allowed options
func ValidateExistence(value string, allowedValues []string, fieldName string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return utils.NewValidationError(
		fmt.Sprintf("%s must be one of: %v", fieldName, allowedValues),
	)
}

// ValidationChain allows chaining multiple validations
type ValidationChain struct {
	errors []error
}

// NewValidationChain creates a new validation chain
func NewValidationChain() *ValidationChain {
	return &ValidationChain{
		errors: make([]error, 0),
	}
}

// Add adds a validation to the chain
func (vc *ValidationChain) Add(validation func() error) *ValidationChain {
	if err := validation(); err != nil {
		vc.errors = append(vc.errors, err)
	}
	return vc
}

// Result returns the first validation error encountered, or nil if all passed
func (vc *ValidationChain) Result() error {
	if len(vc.errors) > 0 {
		return vc.errors[0]
	}
	return nil
}

// AllResults returns all validation errors encountered
func (vc *ValidationChain) AllResults() []error {
	return vc.errors
}

// HasErrors returns true if any validation errors occurred
func (vc *ValidationChain) HasErrors() bool {
	return len(vc.errors) > 0
}
