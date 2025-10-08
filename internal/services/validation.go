package services

import (
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/constants"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// ValidateNotNil checks if an entity is nil and returns a validation error if so
func ValidateNotNil(entity interface{}, entityName string) error {
	if entity == nil {
		return utils.NewValidationError(fmt.Sprintf("%s cannot be nil", entityName))
	}
	return nil
}

// ValidateNotEmpty checks if a string value is empty and returns a validation error if so
func ValidateNotEmpty(value, fieldName string) error {
	if value == "" {
		return utils.NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return nil
}

// ValidatePositive checks if a numeric value is positive and returns a validation error if not
func ValidatePositive(value int64, fieldName string) error {
	if value <= 0 {
		return utils.NewValidationError(fmt.Sprintf("%s must be positive", fieldName))
	}
	return nil
}

// ValidateMinLength checks if a string meets minimum length requirements
func ValidateMinLength(value, fieldName string, minLength int) error {
	if len(value) < minLength {
		return utils.NewValidationError(
			fmt.Sprintf("%s must be at least %d characters", fieldName, minLength),
		)
	}
	return nil
}

// ValidateAPIKey performs comprehensive API key validation
func ValidateAPIKey(apiKey string) error {
	if err := ValidateNotEmpty(apiKey, "api_key"); err != nil {
		return err
	}
	return ValidateMinLength(apiKey, "api_key", constants.MinAPIKeyLength)
}

// ValidateUser performs comprehensive user validation
func ValidateUser(user *models.User) error {
	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(user.Name, "user_name"); err != nil {
		return err
	}

	return ValidateAPIKey(user.ApiKey)
}

// ValidateCookie performs comprehensive cookie validation
func ValidateCookie(cookie *models.Cookie) error {
	if err := ValidateNotNil(cookie, "cookie"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(cookie.Value, "value"); err != nil {
		return err
	}

	if err := ValidateAPIKey(cookie.ApiKey); err != nil {
		return err
	}

	return ValidatePositive(cookie.LastLogin, "last_login")
}

// ValidateNote performs comprehensive note validation
func ValidateNote(note *models.Note) error {
	if err := ValidateNotNil(note, "note"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(note.Content, "content"); err != nil {
		return err
	}

	if note.Level < 0 {
		return utils.NewValidationError("level must be non-negative")
	}

	return nil
}

// ValidateAttachment performs comprehensive attachment validation
func ValidateAttachment(attachment *models.Attachment) error {
	if err := ValidateNotNil(attachment, "attachment"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(attachment.MimeType, "mime_type"); err != nil {
		return err
	}

	if len(attachment.Data) == 0 {
		return utils.NewValidationError("data cannot be empty")
	}

	return nil
}

// ValidateFeed performs comprehensive feed validation
func ValidateFeed(feed *models.Feed) error {
	if err := ValidateNotNil(feed, "feed"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(feed.Title, "title"); err != nil {
		return err
	}

	if err := ValidateNotEmpty(feed.Content, "content"); err != nil {
		return err
	}

	if err := ValidatePositive(feed.UserID, "user_id"); err != nil {
		return err
	}

	return ValidatePositive(feed.CreatedAt, "created_at")
}

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

// ValidateMetalSheet performs comprehensive metal sheet validation
func ValidateMetalSheet(sheet *models.MetalSheet) error {
	if err := ValidateNotNil(sheet, "metal_sheet"); err != nil {
		return err
	}

	if sheet.TileHeight <= 0 {
		return utils.NewValidationError("tile_height must be positive")
	}

	if sheet.Value <= 0 {
		return utils.NewValidationError("value must be positive")
	}

	if sheet.MarkeHeight <= 0 {
		return utils.NewValidationError("marke_height must be positive")
	}

	if sheet.STF <= 0 {
		return utils.NewValidationError("stf must be positive")
	}

	if sheet.STFMax <= 0 {
		return utils.NewValidationError("stf_max must be positive")
	}

	if err := ValidateID(sheet.ToolID, "tool"); err != nil {
		return err
	}

	// Validate machine type identifier
	if !sheet.Identifier.IsValid() {
		return utils.NewValidationError("invalid machine type identifier")
	}

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

// ValidateModificationType validates modification type strings
func ValidateModificationType(modType string) error {
	if err := ValidateNotEmpty(modType, "modification_type"); err != nil {
		return err
	}

	validTypes := []string{
		"trouble_reports",
		"metal_sheets",
		"tools",
		"press_cycles",
		"users",
		"notes",
		"attachments",
	}

	for _, validType := range validTypes {
		if modType == validType {
			return nil
		}
	}

	return utils.NewValidationError(fmt.Sprintf("invalid modification type: %s", modType))
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

// ValidateID checks if an ID is valid (positive)
func ValidateID(id int64, entityName string) error {
	return ValidatePositive(id, fmt.Sprintf("%s_id", entityName))
}

// ValidateTimestamp checks if a timestamp is valid (positive)
func ValidateTimestamp(timestamp int64, fieldName string) error {
	return ValidatePositive(timestamp, fieldName)
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
