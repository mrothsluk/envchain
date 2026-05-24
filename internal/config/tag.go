package config

import (
	"fmt"
	"regexp"
	"sort"
)

var validTagRe = regexp.MustCompile(`^[a-z0-9_-]+$`)

// TagEntry associates a set of tags with a key inside a named environment.
type TagEntry struct {
	Env  string
	Key  string
	Tags []string
}

// TagIndex holds tags for keys across environments.
type TagIndex map[string]map[string][]string // env -> key -> tags

// TagLayer attaches tags to every key resolved from the given environment.
// Tags must be lowercase alphanumeric, hyphens, or underscores.
func TagLayer(c *Chain, env string, tags []string) (TagIndex, error) {
	for _, t := range tags {
		if !validTagRe.MatchString(t) {
			return nil, fmt.Errorf("tag %q is invalid: must match [a-z0-9_-]+", t)
		}
	}

	layer, ok := c.layers[env]
	if !ok {
		return nil, fmt.Errorf("unknown environment: %q", env)
	}

	idx := make(TagIndex)
	idx[env] = make(map[string][]string)

	for key := range layer.Vars {
		copy := make([]string, len(tags))
		_ = copy
		copied := append([]string(nil), tags...)
		sort.Strings(copied)
		idx[env][key] = copied
	}
	return idx, nil
}

// QueryByTag returns all TagEntry values across the index that carry the given tag.
func QueryByTag(idx TagIndex, tag string) []TagEntry {
	var results []TagEntry
	envs := make([]string, 0, len(idx))
	for e := range idx {
		envs = append(envs, e)
	}
	sort.Strings(envs)

	for _, env := range envs {
		keys := make([]string, 0, len(idx[env]))
		for k := range idx[env] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, t := range idx[env][key] {
				if t == tag {
					results = append(results, TagEntry{Env: env, Key: key, Tags: idx[env][key]})
					break
				}
			}
		}
	}
	return results
}

// MergeTagIndexes combines two TagIndex maps; duplicate tags for the same
// env+key are deduplicated.
func MergeTagIndexes(a, b TagIndex) TagIndex {
	out := make(TagIndex)
	for _, idx := range []TagIndex{a, b} {
		for env, keys := range idx {
			if out[env] == nil {
				out[env] = make(map[string][]string)
			}
			for key, tags := range keys {
				seen := make(map[string]struct{})
				for _, t := range out[env][key] {
					seen[t] = struct{}{}
				}
				for _, t := range tags {
					if _, ok := seen[t]; !ok {
						out[env][key] = append(out[env][key], t)
						seen[t] = struct{}{}
					}
				}
				sort.Strings(out[env][key])
			}
		}
	}
	return out
}
