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

type ValidationError struct {
	message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (v *ValidationError) Error() string {
	return "validation error: " + v.message
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

type AlreadyExistsError struct {
	message string
}

func NewAlreadyExistsError(message string) *AlreadyExistsError {
	return &AlreadyExistsError{message: message}
}

func (ae *AlreadyExistsError) Error() string {
	return "already exists: " + ae.message
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

func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if err, ok := err.(*NotFoundError); ok && strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}

	if err, ok := err.(*AlreadyExistsError); ok && strings.Contains(err.Error(), "already exists") {
		return http.StatusConflict
	}

	if err, ok := err.(*InvalidCredentialsError); ok && strings.Contains(err.Error(), "invalid credentials") {
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
