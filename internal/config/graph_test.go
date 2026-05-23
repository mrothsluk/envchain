package config

import (
	"testing"
)

func TestBuildGraphNoReferences(t *testing.T) {
	layer := map[string]string{
		"HOST": "localhost",
		"PORT": "5432",
	}
	g := BuildGraph(layer)
	if len(g.nodes["HOST"]) != 0 {
		t.Errorf("expected no deps for HOST, got %v", g.nodes["HOST"])
	}
	if len(g.nodes["PORT"]) != 0 {
		t.Errorf("expected no deps for PORT, got %v", g.nodes["PORT"])
	}
}

func TestBuildGraphWithReference(t *testing.T) {
	layer := map[string]string{
		"BASE_URL": "http://${HOST}:${PORT}",
		"HOST":     "localhost",
		"PORT":     "8080",
	}
	g := BuildGraph(layer)
	deps := g.nodes["BASE_URL"]
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps for BASE_URL, got %d: %v", len(deps), deps)
	}
}

func TestTopoSortOrdersDepsFirst(t *testing.T) {
	layer := map[string]string{
		"BASE_URL": "http://${HOST}",
		"HOST":     "localhost",
	}
	g := BuildGraph(layer)
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	posHost, posURL := -1, -1
	for i, k := range order {
		if k == "HOST" {
			posHost = i
		}
		if k == "BASE_URL" {
			posURL = i
		}
	}
	if posHost == -1 || posURL == -1 {
		t.Fatal("expected both HOST and BASE_URL in sort order")
	}
	if posHost > posURL {
		t.Errorf("expected HOST before BASE_URL, got positions host=%d url=%d", posHost, posURL)
	}
}

func TestTopoSortDetectsCycle(t *testing.T) {
	// Manually build a cyclic graph
	g := &DependencyGraph{
		nodes: map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {"A"},
		},
	}
	_, err := g.TopoSort()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestCyclesReturnsList(t *testing.T) {
	g := &DependencyGraph{
		nodes: map[string][]string{
			"X": {"Y"},
			"Y": {"X"},
		},
	}
	cycles := g.Cycles()
	if len(cycles) == 0 {
		t.Fatal("expected at least one cycle reported")
	}
}

func TestCyclesEmptyOnAcyclicGraph(t *testing.T) {
	layer := map[string]string{
		"A": "${B}",
		"B": "value",
	}
	g := BuildGraph(layer)
	if c := g.Cycles(); len(c) != 0 {
		t.Errorf("expected no cycles, got %v", c)
	}
}

func TestExtractAllKeysMultiple(t *testing.T) {
	keys := extractAllKeys("${FOO}_${BAR}_${BAZ}")
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}
}

func TestExtractAllKeysNone(t *testing.T) {
	keys := extractAllKeys("no references here")
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %v", keys)
	}
}
