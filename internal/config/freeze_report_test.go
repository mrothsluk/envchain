package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func makeFrozenLayer(env string) *FrozenLayer {
	return &FrozenLayer{
		Env:      env,
		FrozenAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Values: map[string]string{
			"APP_NAME": "envchain",
			"PORT":     "8080",
		},
	}
}

func TestWriteFreezeReportTextContainsEnv(t *testing.T) {
	fl := makeFrozenLayer("staging")
	var buf bytes.Buffer
	if err := WriteFreezeReport(&buf, fl, "text"); err != nil {
		t.Fatalf("WriteFreezeReport: %v", err)
	}
	if !strings.Contains(buf.String(), "staging") {
		t.Errorf("report missing env name: %s", buf.String())
	}
}

func TestWriteFreezeReportTextListsKeys(t *testing.T) {
	fl := makeFrozenLayer("base")
	var buf bytes.Buffer
	_ = WriteFreezeReport(&buf, fl, "text")
	out := buf.String()
	if !strings.Contains(out, "APP_NAME=envchain") {
		t.Errorf("expected APP_NAME in output: %s", out)
	}
	if !strings.Contains(out, "PORT=8080") {
		t.Errorf("expected PORT in output: %s", out)
	}
}

func TestWriteFreezeReportJSONStructure(t *testing.T) {
	fl := makeFrozenLayer("prod")
	var buf bytes.Buffer
	if err := WriteFreezeReport(&buf, fl, "json"); err != nil {
		t.Fatalf("WriteFreezeReport json: %v", err)
	}
	var summary FreezeSummary
	if err := json.Unmarshal(buf.Bytes(), &summary); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if summary.Env != "prod" {
		t.Errorf("env = %q, want prod", summary.Env)
	}
	if summary.Count != 2 {
		t.Errorf("count = %d, want 2", summary.Count)
	}
}

func TestWriteFreezeReportUnknownFormatReturnsError(t *testing.T) {
	fl := makeFrozenLayer("base")
	var buf bytes.Buffer
	err := WriteFreezeReport(&buf, fl, "xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestWriteFreezeReportEmptyDefaultsToText(t *testing.T) {
	fl := makeFrozenLayer("base")
	var buf bytes.Buffer
	if err := WriteFreezeReport(&buf, fl, ""); err != nil {
		t.Fatalf("WriteFreezeReport default: %v", err)
	}
	if !strings.Contains(buf.String(), "Frozen env") {
		t.Errorf("expected text header in output")
	}
}
