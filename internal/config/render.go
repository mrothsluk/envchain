package config

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderOptions controls template rendering behaviour.
type RenderOptions struct {
	RedactSecrets bool
}

// RenderResult holds the output of a template render operation.
type RenderResult struct {
	Output string
	Missing []string
}

// RenderTemplate resolves all variables in the chain for the given environment
// and executes the provided Go template string with those values as data.
// Keys marked as secret are redacted when opts.RedactSecrets is true.
// Missing keys are collected in RenderResult.Missing rather than causing an
// error so the caller can decide how to handle partial renders.
func RenderTemplate(chain *Chain, env, tmplStr string, opts RenderOptions) (RenderResult, error) {
	keys := chain.Keys()
	data := make(map[string]string, len(keys))
	var missing []string

	for _, k := range keys {
		v, err := chain.Resolve(env, k)
		if err != nil {
			missing = append(missing, k)
			data[k] = ""
			continue
		}
		if opts.RedactSecrets && chain.IsSecret(k) {
			data[k] = Redact(v)
		} else {
			data[k] = v
		}
	}

	tmpl, err := template.New("envchain").Option("missingkey=zero").Parse(tmplStr)
	if err != nil {
		return RenderResult{}, fmt.Errorf("render: parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return RenderResult{}, fmt.Errorf("render: execute template: %w", err)
	}

	return RenderResult{
		Output:  buf.String(),
		Missing: missing,
	}, nil
}
