package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment represents a named environment (e.g., dev, staging, prod)
type Environment string

const (
	EnvDev     Environment = "dev"
	EnvStaging Environment = "staging"
	EnvProd    Environment = "prod"
)

// EnvVar holds a single environment variable entry
type EnvVar struct {
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	Secret bool   `yaml:"secret,omitempty"`
}

// EnvChainConfig represents the full layered config file structure
type EnvChainConfig struct {
	Version  int                        `yaml:"version"`
	Layers   map[string][]EnvVar        `yaml:"layers"`
}

// LoadConfig reads and parses an envchain YAML config file
func LoadConfig(path string) (*EnvChainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var cfg EnvChainConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}

	return &cfg, nil
}

// Resolve returns the merged environment variables for the given environment.
// Variables in the target environment override those in the "base" layer.
func (c *EnvChainConfig) Resolve(env Environment) ([]EnvVar, error) {
	resolved := make(map[string]EnvVar)

	for _, v := range c.Layers["base"] {
		resolved[v.Key] = v
	}

	target := string(env)
	if _, ok := c.Layers[target]; !ok && target != "base" {
		return nil, fmt.Errorf("environment %q not found in config", target)
	}

	for _, v := range c.Layers[target] {
		resolved[v.Key] = v
	}

	result := make([]EnvVar, 0, len(resolved))
	for _, v := range resolved {
		result = append(result, v)
	}

	return result, nil
}

// Redact returns the value of an EnvVar, masking it if it is marked secret.
func Redact(v EnvVar) string {
	if v.Secret {
		return strings.Repeat("*", 8)
	}
	return v.Value
}
