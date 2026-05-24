package config

import "fmt"

// RenameResult holds the outcome of a rename operation.
type RenameResult struct {
	OldKey string
	NewKey string
	Env    string
}

// RenameOptions controls how RenameKey behaves on conflicts.
type RenameOptions struct {
	// FailIfExists returns an error if NewKey already exists in the layer.
	FailIfExists bool
	// DryRun reports what would change without mutating the chain.
	DryRun bool
}

// RenameKey renames OldKey to NewKey in the given env layer of chain c.
// The old key is removed and the value is written under the new key.
func RenameKey(c *Chain, env, oldKey, newKey string, opts RenameOptions) (RenameResult, error) {
	if !isValidKey(oldKey) {
		return RenameResult{}, fmt.Errorf("rename: invalid source key %q", oldKey)
	}
	if !isValidKey(newKey) {
		return RenameResult{}, fmt.Errorf("rename: invalid destination key %q", newKey)
	}

	layer, ok := c.layers[env]
	if !ok {
		return RenameResult{}, fmt.Errorf("rename: unknown env %q", env)
	}

	val, exists := layer.values[oldKey]
	if !exists {
		return RenameResult{}, fmt.Errorf("rename: key %q not found in env %q", oldKey, env)
	}

	if _, newExists := layer.values[newKey]; newExists && opts.FailIfExists {
		return RenameResult{}, fmt.Errorf("rename: destination key %q already exists in env %q", newKey, env)
	}

	result := RenameResult{OldKey: oldKey, NewKey: newKey, Env: env}
	if opts.DryRun {
		return result, nil
	}

	layer.values[newKey] = val
	delete(layer.values, oldKey)

	// propagate secret flag
	if layer.secrets[oldKey] {
		layer.secrets[newKey] = true
		delete(layer.secrets, oldKey)
	}

	return result, nil
}
