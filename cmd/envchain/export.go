package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"envchain/internal/config"
)

var exportCmd = &cobra.Command{
	Use:   "export <env>",
	Short: "Export resolved environment variables",
	Long: `Resolve and print environment variables for the given environment.
Supports shell, dotenv, and json output formats.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runExport,
	ValidArgsFunction: completeEnvNames,
}

var (
	exportFormat string
	exportRedact bool
	exportSort   bool
)

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "dotenv",
		"Output format: shell, dotenv, json")
	exportCmd.Flags().BoolVar(&exportRedact, "redact", false,
		"Redact secret values in output")
	exportCmd.Flags().BoolVar(&exportSort, "sort", true,
		"Sort keys alphabetically")
}

func runExport(cmd *cobra.Command, args []string) error {
	env := args[0]

	cfgDir, err := configDir()
	if err != nil {
		return err
	}

	c, err := loadChainFromDir(cfgDir)
	if err != nil {
		return fmt.Errorf("loading chain: %w", err)
	}

	if err := validateExportFormat(exportFormat); err != nil {
		return err
	}

	opts := config.ExportOptions{
		Format:   config.ExportFormat(exportFormat),
		Redact:   exportRedact,
		SortKeys: exportSort,
	}

	if err := config.Export(c, env, opts, os.Stdout); err != nil {
		return fmt.Errorf("export %q: %w", env, err)
	}
	return nil
}

// validateExportFormat checks that the provided format string is one of the
// supported output formats (shell, dotenv, json).
func validateExportFormat(format string) error {
	switch format {
	case "shell", "dotenv", "json":
		return nil
	default:
		return fmt.Errorf("unsupported format %q: must be one of shell, dotenv, json", format)
	}
}
