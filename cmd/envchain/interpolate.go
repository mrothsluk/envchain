package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/example/envchain/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	var (
		files      []string
		envs       []string
		targetEnv  string
		fallbackOS bool
		format     string
	)

	cmd := &cobra.Command{
		Use:   "interpolate",
		Short: "Resolve variable references within a chain for a given environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInterpolate(files, envs, targetEnv, fallbackOS, format)
		},
	}

	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "Layer files in order (base first)")
	cmd.Flags().StringSliceVarP(&envs, "envs", "e", []string{"dev", "staging", "prod"}, "Allowed environments")
	cmd.Flags().StringVarP(&targetEnv, "env", "n", "dev", "Target environment to resolve")
	cmd.Flags().BoolVar(&fallbackOS, "os-fallback", false, "Fall back to OS environment for missing references")
	cmd.Flags().StringVar(&format, "format", "shell", "Output format: shell, dotenv, json")
	_ = cmd.MarkFlagRequired("file")

	rootCmd.AddCommand(cmd)
}

func runInterpolate(files, envs []string, targetEnv string, fallbackOS bool, format string) error {
	if len(files) == 0 {
		return fmt.Errorf("at least one --file is required")
	}

	chain, err := config.NewChain(files[0], envs)
	if err != nil {
		return fmt.Errorf("build chain: %w", err)
	}
	for _, f := range files[1:] {
		if err := chain.AddLayer(targetEnv, f); err != nil {
			return fmt.Errorf("add layer %q: %w", f, err)
		}
	}

	res, err := config.InterpolateChain(chain, targetEnv, fallbackOS)
	if err != nil {
		return err
	}

	if len(res.Missing) > 0 {
		fmt.Fprintf(os.Stderr, "warning: unresolved references: %v\n", res.Missing)
	}

	keys := make([]string, 0, len(res.Values))
	for k := range res.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(res.Values)
	case "dotenv":
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "%s=%q\n", k, res.Values[k])
		}
	default: // shell
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "export %s=%q\n", k, res.Values[k])
		}
	}
	return nil
}
