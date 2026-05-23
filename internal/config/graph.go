package config

import (
	"fmt"
	"sort"
	"strings"
)

// DependencyGraph represents key-level dependencies derived from interpolation
// references within a single environment layer.
type DependencyGraph struct {
	nodes map[string][]string // key -> keys it depends on
}

// BuildGraph inspects a layer's values for ${KEY} references and constructs
// a directed dependency graph. Unknown reference targets are included as nodes
// with no outgoing edges.
func BuildGraph(layer map[string]string) *DependencyGraph {
	g := &DependencyGraph{nodes: make(map[string][]string)}
	for key, val := range layer {
		if _, ok := g.nodes[key]; !ok {
			g.nodes[key] = []string{}
		}
		refs := extractAllKeys(val)
		for _, ref := range refs {
			if _, ok := g.nodes[ref]; !ok {
				g.nodes[ref] = []string{}
			}
			g.nodes[key] = append(g.nodes[key], ref)
		}
	}
	return g
}

// TopoSort returns keys in topological order (dependencies before dependents).
// Returns an error if a cycle is detected.
func (g *DependencyGraph) TopoSort() ([]string, error) {
	visited := map[string]int{} // 0=unvisited,1=in-progress,2=done
	var order []string

	keys := make([]string, 0, len(g.nodes))
	for k := range g.nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var visit func(n string) error
	visit = func(n string) error {
		switch visited[n] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("cycle detected at key %q", n)
		}
		visited[n] = 1
		for _, dep := range g.nodes[n] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visited[n] = 2
		order = append(order, n)
		return nil
	}

	for _, k := range keys {
		if err := visit(k); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// Cycles returns all detected cycles as human-readable strings.
func (g *DependencyGraph) Cycles() []string {
	_, err := g.TopoSort()
	if err != nil {
		return []string{err.Error()}
	}
	return nil
}

// extractAllKeys returns all ${KEY} references found in val.
func extractAllKeys(val string) []string {
	var keys []string
	for {
		start := strings.Index(val, "${")
		if start == -1 {
			break
		}
		end := strings.Index(val[start:], "}")
		if end == -1 {
			break
		}
		key := val[start+2 : start+end]
		if key != "" {
			keys = append(keys, key)
		}
		val = val[start+end+1:]
	}
	return keys
}
