package config

import (
	"strings"
	"testing"
)

func TestValidateLayerValidKeys(t *testing.T) {
	env := map[string]string{
		"APP_HOST": "localhost",
		"PORT":     "8080",
	}
	if err := ValidateLayer(env, nil); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateLayerInvalidKeyLowercase(t *testing.T) {
	env := map[string]string{
		"app_host": "localhost",
	}
	err := ValidateLayer(env, nil)
	if err == nil {
		t.Fatal("expected error for lowercase key, got nil")
	}
	if !strings.Contains(err.Error(), "app_host") {
		t.Errorf("error should mention the bad key, got: %v", err)
	}
}

func TestValidateLayerEmptyValue(t *testing.T) {
	env := map[string]string{
		"APP_HOST": "   ",
	}
	err := ValidateLayer(env, nil)
	if err == nil {
		t.Fatal("expected error for empty value, got nil")
	}
	if !strings.Contains(err.Error(), "empty or whitespace") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateLayerMissingRequired(t *testing.T) {
	env := map[string]string{
		"APP_HOST": "localhost",
	}
	err := ValidateLayer(env, []string{"APP_HOST", "DB_URL"})
	if err == nil {
		t.Fatal("expected error for missing required key, got nil")
	}
	if !strings.Contains(err.Error(), "DB_URL") {
		t.Errorf("error should mention missing key DB_URL, got: %v", err)
	}
}

func TestValidateLayerMultipleErrors(t *testing.T) {
	env := map[string]string{
		"bad-key": "",
	}
	err := ValidateLayer(env, nil)
	if err == nil {
		t.Fatal("expected errors, got nil")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidateChainRequiredKeyPresent(t *testing.T) {
	c := buildTestChain()
	if err := ValidateChain(c, "dev", []string{"APP_HOST"}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateChainRequiredKeyMissing(t *testing.T) {
	c := buildTestChain()
	err := ValidateChain(c, "dev", []string{"NONEXISTENT_KEY"})
	if err == nil {
		t.Fatal("expected error for missing chain key, got nil")
	}
	if !strings.Contains(err.Error(), "NONEXISTENT_KEY") {
		t.Errorf("error should mention missing key, got: %v", err)
	}
}

func TestValidateChainNilChain(t *testing.T) {
	err := ValidateChain(nil, "dev", nil)
	if err == nil {
		t.Fatal("expected error for nil chain, got nil")
	}
}
