package config

import "fmt"

// CloneOptions controls how a layer is cloned into a new environment.
type CloneOptions struct {
	// SkipSecrets omits keys marked as secret from the clone.
	SkipSecrets bool
	// OnlyKeys restricts the clone to the listed keys; empty means all keys.
	OnlyKeys []string
}

// CloneLayer copies all resolved values from srcEnv into dstEnv within chain,
// applying the provided options. The destination environment must already exist
// as a layer in the chain.
func CloneLayer(chain *Chain, srcEnv, dstEnv string, opts CloneOptions) (int, error) {
	srcLayer, ok := chain.layers[srcEnv]
	if !ok {
		return 0, fmt.Errorf("clone: unknown source environment %q", srcEnv)
	}
	dstLayer, ok := chain.layers[dstEnv]
	if !ok {
		return 0, fmt.Errorf("clone: unknown destination environment %q", dstEnv)
	}

	allowSet := make(map[string]bool, len(opts.OnlyKeys))
	for _, k := range opts.OnlyKeys {
		allowSet[k] = true
	}

	copied := 0
	for key, entry := range srcLayer.entries {
		if len(allowSet) > 0 && !allowSet[key] {
			continue
		}
		if opts.SkipSecrets && entry.Secret {
			continue
		}
		dstLayer.entries[key] = entry
		copied++
	}
	return copied, nil
}
