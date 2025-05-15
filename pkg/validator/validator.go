package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// Errors represents validation errors
type Errors map[string]string

// Validator provides methods for validating data
type Validator struct {
	Errors Errors
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		Errors: make(Errors),
	}
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error for a specific field
func (v *Validator) AddError(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

// Check adds an error if the condition is false
func (v *Validator) Check(condition bool, field, message string) {
	if !condition {
		v.AddError(field, message)
	}
}

// NotEmpty checks if a string value is not empty
func (v *Validator) NotEmpty(value string, field string) {
	v.Check(value != "", field, "must not be empty")
}

// MaxLength checks if a string value does not exceed the maximum length
func (v *Validator) MaxLength(value string, field string, maxLength int) {
	v.Check(len(value) <= maxLength, field, fmt.Sprintf("must not exceed %d characters", maxLength))
}

// MinLength checks if a string value is at least the minimum length
func (v *Validator) MinLength(value string, field string, minLength int) {
	v.Check(len(value) >= minLength, field, fmt.Sprintf("must be at least %d characters", minLength))
}

// Matches checks if a string value matches a regular expression
func (v *Validator) Matches(value string, field string, pattern *regexp.Regexp) {
	v.Check(pattern.MatchString(value), field, "invalid format")
}

// NoSpecialChars checks if a string value has no special characters
func (v *Validator) NoSpecialChars(value string, field string) {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	v.Check(pattern.MatchString(value), field, "must not contain special characters")
}

// NoPathTraversal checks if a string value has no path traversal characters
func (v *Validator) NoPathTraversal(value string, field string) {
	v.Check(!strings.Contains(value, ".."), field, "must not contain path traversal characters")
}

// GetErrors returns all errors
func (v *Validator) GetErrors() Errors {
	return v.Errors
}
