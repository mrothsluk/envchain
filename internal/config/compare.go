package config

import (
	"fmt"
	"sort"
)

// CompareResult holds the outcome of comparing two environments.
type CompareResult struct {
	EnvA      string
	EnvB      string
	OnlyInA   map[string]string
	OnlyInB   map[string]string
	Different map[string][2]string // key -> [valueA, valueB]
	Identical []string
}

// SortedOnlyInA returns keys present only in EnvA, sorted.
func (r *CompareResult) SortedOnlyInA() []string {
	return sortedKeys(r.OnlyInA)
}

// SortedOnlyInB returns keys present only in EnvB, sorted.
func (r *CompareResult) SortedOnlyInB() []string {
	return sortedKeys(r.OnlyInB)
}

// SortedDifferent returns keys with differing values, sorted.
func (r *CompareResult) SortedDifferent() []string {
	keys := make([]string, 0, len(r.Different))
	for k := range r.Different {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// CompareEnvs resolves both named environments from the chain and compares
// their key/value pairs, redacting secrets in the result.
func CompareEnvs(chain *Chain, envA, envB string, redact bool) (*CompareResult, error) {
	valuesA, err := resolveAll(chain, envA)
	if err != nil {
		return nil, fmt.Errorf("compare: resolve %q: %w", envA, err)
	}
	valuesB, err := resolveAll(chain, envB)
	if err != nil {
		return nil, fmt.Errorf("compare: resolve %q: %w", envB, err)
	}

	if redact {
		valuesA = redactMap(chain, envA, valuesA)
		valuesB = redactMap(chain, envB, valuesB)
	}

	result := &CompareResult{
		EnvA:      envA,
		EnvB:      envB,
		OnlyInA:   make(map[string]string),
		OnlyInB:   make(map[string]string),
		Different: make(map[string][2]string),
	}

	allKeys := unionKeys(valuesA, valuesB)
	for _, k := range allKeys {
		va, inA := valuesA[k]
		vb, inB := valuesB[k]
		switch {
		case inA && !inB:
			result.OnlyInA[k] = va
		case !inA && inB:
			result.OnlyInB[k] = vb
		case va == vb:
			result.Identical = append(result.Identical, k)
		default:
			result.Different[k] = [2]string{va, vb}
		}
	}
	sort.Strings(result.Identical)
	return result, nil
}

func redactMap(chain *Chain, env string, values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		out[k] = Redact(chain, env, k, v)
	}
	return out
}
