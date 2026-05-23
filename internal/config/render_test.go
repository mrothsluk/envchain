package config

import (
	"strings"
	"testing"
)

func buildRenderChain(t *testing.T) *Chain {
	t.Helper()

	base := writeTemp(t, `
APP_NAME=envchain
DB_HOST=localhost
DB_PASS=secret123
PORT=8080
`)
	prod := writeTemp(t, `
DB_HOST=prod.db.internal
DB_PASS=supersecret
`)

	cfgBase, err := LoadConfig(base)
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	cfgBase.Secrets = []string{"DB_PASS"}

	cfgProd, err := LoadConfig(prod)
	if err != nil {
		t.Fatalf("load prod: %v", err)
	}

	chain := NewChain()
	chain.AddLayer("base", cfgBase)
	chain.AddLayer("prod", cfgProd)
	return chain
}

func TestRenderTemplateBase(t *testing.T) {
	chain := buildRenderChain(t)
	tmpl := "host={{.DB_HOST}} port={{.PORT}}"
	res, err := RenderTemplate(chain, "base", tmpl, RenderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "host=localhost port=8080" {
		t.Errorf("got %q", res.Output)
	}
	if len(res.Missing) != 0 {
		t.Errorf("expected no missing keys, got %v", res.Missing)
	}
}

func TestRenderTemplateProdOverrides(t *testing.T) {
	chain := buildRenderChain(t)
	tmpl := "host={{.DB_HOST}}"
	res, err := RenderTemplate(chain, "prod", tmpl, RenderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "host=prod.db.internal" {
		t.Errorf("got %q", res.Output)
	}
}

func TestRenderTemplateRedactsSecrets(t *testing.T) {
	chain := buildRenderChain(t)
	tmpl := "pass={{.DB_PASS}}"
	res, err := RenderTemplate(chain, "base", tmpl, RenderOptions{RedactSecrets: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(res.Output, "secret123") {
		t.Errorf("secret value leaked in output: %q", res.Output)
	}
	if !strings.Contains(res.Output, "***") {
		t.Errorf("expected redaction marker in output: %q", res.Output)
	}
}

func TestRenderTemplateMissingKeyCollected(t *testing.T) {
	chain := buildRenderChain(t)
	// UNKNOWN_KEY is not in the chain; expect it in Missing, not an error.
	tmpl := "app={{.APP_NAME}} unknown={{.UNKNOWN_KEY}}"
	res, err := RenderTemplate(chain, "base", tmpl, RenderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Missing) == 0 {
		t.Error("expected UNKNOWN_KEY in Missing")
	}
}

func TestRenderTemplateInvalidTemplateSyntax(t *testing.T) {
	chain := buildRenderChain(t)
	_, err := RenderTemplate(chain, "base", "{{.UNCLOSED", RenderOptions{})
	if err == nil {
		t.Error("expected parse error for invalid template")
	}
}
