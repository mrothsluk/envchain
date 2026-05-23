package config

import "fmt"

// Entry holds a single environment variable value and its metadata.
type Entry struct {
	Value  string
	Secret bool
}

// layer holds the entries for one named environment.
type layer struct {
	name    string
	entries map[string]Entry
}

// Chain is an ordered stack of named environment layers.
type Chain struct {
	order  []string
	layers map[string]*layer
}

// NewChain returns an empty Chain.
func NewChain() *Chain {
	return &Chain{layers: make(map[string]*layer)}
}

// AddLayerEntries registers a named layer with the provided entries.
func (c *Chain) AddLayerEntries(name string, entries map[string]Entry) {
	c.order = append(c.order, name)
	c.layers[name] = &layer{name: name, entries: entries}
}

// Resolve returns the effective Entry for key in the given environment by
// walking layers in order and returning the last value that matches.
func (c *Chain) Resolve(env, key string) (Entry, error) {
	if _, ok := c.layers[env]; !ok {
		return Entry{}, fmt.Errorf("chain: unknown environment %q", env)
	}
	var found *Entry
	for _, name := range c.order {
		l := c.layers[name]
		if e, ok := l.entries[key]; ok {
			e2 := e
			found = &e2
		}
		if name == env {
			break
		}
	}
	if found == nil {
		return Entry{}, fmt.Errorf("chain: key %q not found in environment %q", key, env)
	}
	return *found, nil
}

// Envs returns the ordered list of environment names.
func (c *Chain) Envs() []string {
	out := make([]string, len(c.order))
	copy(out, c.order)
	return out
}

// Keys returns the union of all keys visible in the given environment.
func (c *Chain) Keys(env string) ([]string, error) {
	if _, ok := c.layers[env]; !ok {
		return nil, fmt.Errorf("chain: unknown environment %q", env)
	}
	seen := make(map[string]struct{})
	for _, name := range c.order {
		for k := range c.layers[name].entries {
			seen[k] = struct{}{}
		}
		if name == env {
			break
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	return keys, nil
}
