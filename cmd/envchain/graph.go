package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"envchain/internal/config"
)

var graphFormat string

func init() {
	graphCmd := &cobra.Command{
		Use:   "graph <file>",
		Short: "Show key dependency graph for an env file",
		Args:  cobra.ExactArgs(1),
		RunE:  runGraph,
	}
	graphCmd.Flags().StringVarP(&graphFormat, "format", "f", "text", "Output format: text|json")
	rootCmd.AddCommand(graphCmd)
}

func runGraph(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	g := config.BuildGraph(cfg.Values)

	cycles := g.Cycles()
	if len(cycles) > 0 {
		fmt.Fprintln(os.Stderr, "WARNING: cycles detected:")
		for _, c := range cycles {
			fmt.Fprintln(os.Stderr, " ", c)
		}
	}

	order, err := g.TopoSort()
	if err != nil {
		return fmt.Errorf("topo sort: %w", err)
	}

	switch strings.ToLower(graphFormat) {
	case "json":
		return printGraphJSON(g, order)
	default:
		printGraphText(g, order)
		return nil
	}
}

func printGraphText(g *config.DependencyGraph, order []string) {
	fmt.Println("Topological order (dependencies first):")
	for i, k := range order {
		fmt.Printf("  %d. %s\n", i+1, k)
	}
	fmt.Println()
	fmt.Println("Edges (key -> depends on):")
	for _, k := range order {
		deps := g.Nodes()[k]
		if len(deps) > 0 {
			fmt.Printf("  %s -> [%s]\n", k, strings.Join(deps, ", "))
		}
	}
}

func printGraphJSON(g *config.DependencyGraph, order []string) error {
	out := map[string]interface{}{
		"order": order,
		"edges": g.Nodes(),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
