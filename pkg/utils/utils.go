package utils

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

type DatabaseError struct {
	message string
	table   string
	err     error
}

func NewDatabaseError(action, table string, err error) *DatabaseError {
	return &DatabaseError{message: action + " error", table: table, err: err}
}

func (de *DatabaseError) Error() string {
	return de.message + ": " + de.table + ": " + de.err.Error()
}

func IsNotDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*DatabaseError); ok {
		return true
	}

	return false
}

type ValidationError struct {
	message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (v *ValidationError) Error() string {
	return "validation error: " + v.message
}

func IsNotValidationError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*ValidationError); ok {
		return true
	}

	return false
}

type NotFoundError struct {
	message string
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message: message}
}

func (nf *NotFoundError) Error() string {
	return "not found: " + nf.message
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*NotFoundError); ok {
		return true
	}

	return false
}

type AlreadyExistsError struct {
	message string
}

func NewAlreadyExistsError(message string) *AlreadyExistsError {
	return &AlreadyExistsError{message: message}
}

func (ae *AlreadyExistsError) Error() string {
	return "already exists: " + ae.message
}

func IsAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*AlreadyExistsError); ok {
		return true
	}

	return false
}

type InvalidCredentialsError struct {
	message string
}

func NewInvalidCredentialsError(message string) *InvalidCredentialsError {
	return &InvalidCredentialsError{message: message}
}

func (ic *InvalidCredentialsError) Error() string {
	return "invalid credentials: " + ic.message
}

func IsInvalidCredentialsError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*InvalidCredentialsError); ok {
		return true
	}

	return false
}

func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if IsNotFoundError(err) {
		return http.StatusNotFound
	}

	if IsAlreadyExistsError(err) {
		return http.StatusConflict
	}

	if IsInvalidCredentialsError(err) {
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}

// MaskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
func MaskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
