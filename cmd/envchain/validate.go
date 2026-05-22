package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"envchain/internal/config"
)

var (
	flagValidateEnv      string
	flagValidateRequired string
)

func init() {
	validateCmd := &cobra.Command{
		Use:   "validate <config-file>",
		Short: "Validate a config file for a given environment",
		Args:  cobra.ExactArgs(1),
		RunE:  runValidate,
	}

	validateCmd.Flags().StringVarP(&flagValidateEnv, "env", "e", "dev",
		"target environment (dev, staging, prod)")
	validateCmd.Flags().StringVarP(&flagValidateRequired, "required", "r", "",
		"comma-separated list of required keys")

	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfgFile := args[0]

	layers, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config %q: %w", cfgFile, err)
	}

	chain := config.NewChain()
	for _, layer := range layers {
		if err := chain.AddLayer(layer); err != nil {
			return fmt.Errorf("building chain: %w", err)
		}
	}

	var required []string
	if flagValidateRequired != "" {
		for _, k := range strings.Split(flagValidateRequired, ",") {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				required = append(required, trimmed)
			}
		}
	}

	if err := config.ValidateChain(chain, flagValidateEnv, required); err != nil {
		fmt.Fprintf(os.Stderr, "envchain: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ config %q is valid for environment %q\n", cfgFile, flagValidateEnv)
	return nil
}
