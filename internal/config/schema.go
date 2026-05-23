package config

import (
	"fmt"
	"regexp"
	"strings"
)

// SchemaField describes a single expected environment variable.
type SchemaField struct {
	Key      string
	Required bool
	Pattern  string // optional regex the value must match
	Secret   bool
}

// Schema holds the declared fields for a layer.
type Schema struct {
	Fields []SchemaField
	index  map[string]SchemaField
}

// NewSchema builds a Schema and validates that all keys are well-formed.
func NewSchema(fields []SchemaField) (*Schema, error) {
	idx := make(map[string]SchemaField, len(fields))
	for _, f := range fields {
		if !validKeyRe.MatchString(f.Key) {
			return nil, fmt.Errorf("schema: invalid key %q", f.Key)
		}
		if f.Pattern != "" {
			if _, err := regexp.Compile(f.Pattern); err != nil {
				return nil, fmt.Errorf("schema: key %q has invalid pattern: %w", f.Key, err)
			}
		}
		idx[f.Key] = f
	}
	return &Schema{Fields: fields, index: idx}, nil
}

// validKeyRe mirrors the key rules used elsewhere in the project.
var validKeyRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// SchemaError represents a single schema violation.
type SchemaError struct {
	Key     string
	Message string
}

func (e SchemaError) Error() string {
	return fmt.Sprintf("key %q: %s", e.Key, e.Message)
}

// ValidateAgainstSchema checks that the resolved values for env satisfy s.
// It returns all violations found, not just the first.
func ValidateAgainstSchema(s *Schema, values map[string]string) []SchemaError {
	var errs []SchemaError

	for _, f := range s.Fields {
		v, ok := values[f.Key]
		if !ok || strings.TrimSpace(v) == "" {
			if f.Required {
				errs = append(errs, SchemaError{Key: f.Key, Message: "required key is missing or empty"})
			}
			continue
		}
		if f.Pattern != "" {
			re := regexp.MustCompile(f.Pattern)
			if !re.MatchString(v) {
				errs = append(errs, SchemaError{
					Key:     f.Key,
					Message: fmt.Sprintf("value does not match pattern %q", f.Pattern),
				})
			}
		}
	}
	return errs
}

// SchemaKeys returns the set of keys declared in the schema.
func (s *Schema) SchemaKeys() []string {
	keys := make([]string, 0, len(s.Fields))
	for _, f := range s.Fields {
		keys = append(keys, f.Key)
	}
	return keys
}
