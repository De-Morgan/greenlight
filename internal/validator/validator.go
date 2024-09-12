package validator

import (
	"regexp"
	"slices"
)

// Validator type which contains a map of validation errors.
type Validator struct{ Errors map[string]string }

// New is a helper which creates a new Validator instance with an empty errors map.
func New() *Validator { return &Validator{Errors: make(map[string]string)} }

// Valid returns true if the errors map doesn't contain any entries.
func (v *Validator) Valid() bool { return len(v.Errors) == 0 }

// AddError adds an error message to the map
// it doesn't override existing errors
func (v *Validator) AddError(key, message string) {
	if v == nil {
		return
	}
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map only if a validation check is invalid.
func (v *Validator) Check(invalid bool, key, message string) {
	if v == nil {
		return
	}
	if invalid {
		v.AddError(key, message)
	}
}

// Matches returns true if a string value matches a specific regexp pattern.
func Matches(value string, rx *regexp.Regexp) bool { return rx.MatchString(value) }

// Unique returns true if all values in a slice are unique.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}

// In returns true if a specific value is in a list of values.
func In[E comparable](value E, values ...E) bool {
	return slices.Contains(values, value)
}
