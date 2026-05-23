package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/envchain/internal/config"
)

func init() {
	var env string

	cmd := &cobra.Command{
		Use:   "lint [files...]",
		Short: "Lint environment config files for common issues",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(args, env)
		},
	}
	cmd.Flags().StringVar(&env, "env", "dev", "target environment for chain resolution")
	rootCmd.AddCommand(cmd)
}

func runLint(files []string, env string) error {
	chain, err := buildChainFromFiles(files, env)
	if err != nil {
		return fmt.Errorf("building chain: %w", err)
	}

	result := config.LintChain(chain)

	if len(result.Findings) == 0 {
		fmt.Println("✓ No lint issues found.")
		return nil
	}

	for _, f := range result.Findings {
		fmt.Fprintln(os.Stderr, f.String())
	}

	if result.HasErrors() {
		return fmt.Errorf("lint failed with %d error(s)", countBySeverity(result, config.LintError))
	}

	fmt.Fprintf(os.Stderr, "lint: %d warning(s)\n", countBySeverity(result, config.LintWarn))
	return nil
}

func countBySeverity(result config.LintResult, sev config.LintSeverity) int {
	n := 0
	for _, f := range result.Findings {
		if f.Severity == sev {
			n++
		}
	}
	return n
}

func buildChainFromFiles(files []string, env string) (*config.Chain, error) {
	chain := config.NewChain()
	for _, f := range files {
		layer, err := config.LoadConfig(f)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", f, err)
		}
		if err := chain.AddLayer(f, layer, env); err != nil {
			return nil, fmt.Errorf("adding layer %s: %w", f, err)
		}
	}
	return chain, nil
}
