package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ValidationError holds all validation issues found in a config layer.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d error(s):\n  - %s",
		len(e.Errors), strings.Join(e.Errors, "\n  - "))
}

var validKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// ValidateLayer checks that all keys and values in an env map satisfy
// envchain naming conventions and that required keys are present.
func ValidateLayer(env map[string]string, required []string) error {
	var errs []string

	for k, v := range env {
		if !validKeyPattern.MatchString(k) {
			errs = append(errs, fmt.Sprintf("key %q must match pattern [A-Z][A-Z0-9_]*", k))
		}
		if strings.TrimSpace(v) == "" {
			errs = append(errs, fmt.Sprintf("key %q has an empty or whitespace-only value", k))
		}
	}

	for _, req := range required {
		if _, ok := env[req]; !ok {
			errs = append(errs, fmt.Sprintf("required key %q is missing", req))
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// ValidateChain resolves every key in the chain and verifies required keys
// are reachable for the given environment.
func ValidateChain(c *Chain, env string, required []string) error {
	if c == nil {
		return errors.New("chain must not be nil")
	}

	var errs []string
	for _, req := range required {
		if _, err := c.Resolve(env, req); err != nil {
			errs = append(errs, fmt.Sprintf("required key %q not resolvable in env %q: %v", req, env, err))
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}
