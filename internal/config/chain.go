package config

import (
	"fmt"
	"maps"
)

// Env represents a named environment tier.
type Env string

const (
	EnvBase    Env = "base"
	EnvDev     Env = "dev"
	EnvStaging Env = "staging"
	EnvProd    Env = "prod"
)

// envOrder defines the precedence chain: later envs override earlier ones.
var envOrder = []Env{EnvBase, EnvDev, EnvStaging, EnvProd}

// Chain holds configs for each environment tier.
type Chain struct {
	layers map[Env]map[string]string
}

// NewChain creates an empty Chain.
func NewChain() *Chain {
	return &Chain{layers: make(map[Env]map[string]string)}
}

// AddLayer registers a config map for the given environment tier.
func (c *Chain) AddLayer(env Env, values map[string]string) error {
	for _, valid := range envOrder {
		if env == valid {
			c.layers[env] = values
			return nil
		}
	}
	return fmt.Errorf("unknown environment tier: %q", env)
}

// Resolve merges all layers up to and including the target env,
// with later tiers taking precedence over earlier ones.
func (c *Chain) Resolve(target Env) (map[string]string, error) {
	found := false
	for _, e := range envOrder {
		if e == target {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown environment tier: %q", target)
	}

	result := make(map[string]string)
	for _, e := range envOrder {
		if layer, ok := c.layers[e]; ok {
			maps.Copy(result, layer)
		}
		if e == target {
			break
		}
	}
	return result, nil
}
