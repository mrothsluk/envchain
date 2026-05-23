package config

import (
	"testing"
)

func TestNewSchemaRejectsLowercaseKey(t *testing.T) {
	_, err := NewSchema([]SchemaField{{Key: "bad_key", Required: true}})
	if err == nil {
		t.Fatal("expected error for lowercase key, got nil")
	}
}

func TestNewSchemaRejectsInvalidPattern(t *testing.T) {
	_, err := NewSchema([]SchemaField{{Key: "PORT", Pattern: "[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}
}

func TestValidateAgainstSchemaMissingRequired(t *testing.T) {
	s, err := NewSchema([]SchemaField{
		{Key: "DATABASE_URL", Required: true},
	})
	if err != nil {
		t.Fatalf("NewSchema: %v", err)
	}
	errs := ValidateAgainstSchema(s, map[string]string{})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Key != "DATABASE_URL" {
		t.Errorf("unexpected key %q", errs[0].Key)
	}
}

func TestValidateAgainstSchemaPatternMismatch(t *testing.T) {
	s, err := NewSchema([]SchemaField{
		{Key: "PORT", Required: true, Pattern: `^\d+$`},
	})
	if err != nil {
		t.Fatalf("NewSchema: %v", err)
	}
	errs := ValidateAgainstSchema(s, map[string]string{"PORT": "not-a-number"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateAgainstSchemaPatternMatch(t *testing.T) {
	s, err := NewSchema([]SchemaField{
		{Key: "PORT", Required: true, Pattern: `^\d+$`},
	})
	if err != nil {
		t.Fatalf("NewSchema: %v", err)
	}
	errs := ValidateAgainstSchema(s, map[string]string{"PORT": "8080"})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateAgainstSchemaOptionalMissingIsOK(t *testing.T) {
	s, err := NewSchema([]SchemaField{
		{Key: "LOG_LEVEL", Required: false},
	})
	if err != nil {
		t.Fatalf("NewSchema: %v", err)
	}
	errs := ValidateAgainstSchema(s, map[string]string{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors for optional missing key, got %v", errs)
	}
}

func TestSchemaKeysReturnsAll(t *testing.T) {
	s, _ := NewSchema([]SchemaField{
		{Key: "A"},
		{Key: "B"},
		{Key: "C"},
	})
	keys := s.SchemaKeys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
}

func TestValidateAgainstSchemaMultipleErrors(t *testing.T) {
	s, _ := NewSchema([]SchemaField{
		{Key: "A", Required: true},
		{Key: "B", Required: true},
	})
	errs := ValidateAgainstSchema(s, map[string]string{})
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}
