package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/envchain/internal/config"
	"github.com/spf13/cobra"
)

var transformCmd = &cobra.Command{
	Use:   "transform <env> <rule>[,rule...]",
	Short: "Apply value transformations to an environment layer",
	Args:  cobra.ExactArgs(2),
	RunE:  runTransform,
}

func init() {
	transformCmd.Flags().StringSliceP("files", "f", nil, "config files to load (base first)")
	transformCmd.Flags().BoolP("dry-run", "n", false, "preview changes without applying them")
	transformCmd.Flags().StringP("format", "o", "text", "output format: text|json")
	rootCmd.AddCommand(transformCmd)
}

func runTransform(cmd *cobra.Command, args []string) error {
	env := args[0]
	rules := strings.Split(args[1], ",")

	files, _ := cmd.Flags().GetStringSlice("files")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	format, _ := cmd.Flags().GetString("format")

	if len(files) == 0 {
		return fmt.Errorf("at least one --files entry is required")
	}

	c := config.NewChain()
	for _, f := range files {
		cfg, err := config.LoadConfig(f)
		if err != nil {
			return fmt.Errorf("loading %s: %w", f, err)
		}
		if err := c.AddLayer(env, cfg); err != nil {
			return fmt.Errorf("adding layer %s: %w", f, err)
		}
	}

	results, err := config.TransformLayer(c, env, rules, dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintln(os.Stderr, "[dry-run] no changes written")
	}

	return config.WriteTransformReport(os.Stdout, results, format)
}
