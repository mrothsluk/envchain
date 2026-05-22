package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourorg/envchain/internal/config"
)

var diffCmd = &cobra.Command{
	Use:   "diff <env1> <env2> [config-file]",
	Short: "Show differences in resolved values between two environments",
	Long: `Compare the resolved environment variable values between two environments.

Displays keys that have changed, been added, or removed when moving from
env1 to env2. Secret values are redacted in the output.

Examples:
  envchain diff dev prod
  envchain diff dev staging config/envchain.yaml
  envchain diff staging prod --show-unchanged`,
	Args: cobra.RangeArgs(2, 3),
	RunE: runDiff,
}

var showUnchanged bool

func init() {
	diffCmd.Flags().BoolVar(&showUnchanged, "show-unchanged", false, "Also display keys whose values did not change")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	env1 := args[0]
	env2 := args[1]

	cfgFile := "envchain.yaml"
	if len(args) == 3 {
		cfgFile = args[2]
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	chain1, err := config.NewChain(cfg, env1)
	if err != nil {
		return fmt.Errorf("building chain for %q: %w", env1, err)
	}

	chain2, err := config.NewChain(cfg, env2)
	if err != nil {
		return fmt.Errorf("building chain for %q: %w", env2, err)
	}

	results, err := config.DiffChain(chain1, chain2)
	if err != nil {
		return fmt.Errorf("computing diff: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stdout, "No differences between %s and %s\n", env1, env2)
		return nil
	}

	fmt.Fprintf(os.Stdout, "Diff: %s → %s\n", env1, env2)
	fmt.Fprintln(os.Stdout, strings.Repeat("-", 60))

	changed := 0
	for _, r := range results {
		switch {
		case r.Added:
			fmt.Fprintf(os.Stdout, "+ %-30s %s\n", r.Key, r.NewValue)
			changed++
		case r.Removed:
			fmt.Fprintf(os.Stdout, "- %-30s %s\n", r.Key, r.OldValue)
			changed++
		case r.Changed:
			fmt.Fprintf(os.Stdout, "~ %-30s %s → %s\n", r.Key, r.OldValue, r.NewValue)
			changed++
		default:
			if showUnchanged {
				fmt.Fprintf(os.Stdout, "  %-30s %s\n", r.Key, r.NewValue)
			}
		}
	}

	fmt.Fprintln(os.Stdout, strings.Repeat("-", 60))
	fmt.Fprintf(os.Stdout, "%d key(s) changed\n", changed)

	return nil
}
