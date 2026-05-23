package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/your-org/envchain/internal/config"
	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render <env> <template-string>",
	Short: "Render a Go template using resolved environment variables",
	Args:  cobra.ExactArgs(2),
	RunE:  runRender,
}

var renderFiles []string
var renderRedact bool

func init() {
	renderCmd.Flags().StringArrayVarP(&renderFiles, "file", "f", nil, "Config files to load (base first)")
	renderCmd.Flags().BoolVar(&renderRedact, "redact", false, "Redact secret values in output")
	rootCmd.AddCommand(renderCmd)
}

func runRender(cmd *cobra.Command, args []string) error {
	env := args[0]
	tmplStr := args[1]

	if len(renderFiles) == 0 {
		return fmt.Errorf("at least one --file is required")
	}

	chain := config.NewChain()
	for _, f := range renderFiles {
		cfg, err := config.LoadConfig(f)
		if err != nil {
			return fmt.Errorf("load %s: %w", f, err)
		}
		chain.AddLayer(f, cfg)
	}

	opts := config.RenderOptions{RedactSecrets: renderRedact}
	res, err := config.RenderTemplate(chain, env, tmplStr, opts)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	if len(res.Missing) > 0 {
		fmt.Fprintf(os.Stderr, "warn: missing keys: %s\n", strings.Join(res.Missing, ", "))
	}

	fmt.Fprintln(cmd.OutOrStdout(), res.Output)
	return nil
}
