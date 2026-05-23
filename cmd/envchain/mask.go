package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/yourorg/envchain/internal/config"
)

var maskCmd = &cobra.Command{
	Use:   "mask <env> <file...>",
	Short: "Apply masking rules to environment variable values",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runMask,
}

var (
	maskFormat      string
	maskSecrets     bool
	maskPatterns    []string
	maskReplacement string
)

func init() {
	maskCmd.Flags().StringVarP(&maskFormat, "format", "f", "text", "Output format: text|json")
	maskCmd.Flags().BoolVar(&maskSecrets, "secrets", true, "Fully redact keys marked as secrets")
	maskCmd.Flags().StringArrayVar(&maskPatterns, "pattern", nil, "Regex pattern to mask (may repeat)")
	maskCmd.Flags().StringVar(&maskReplacement, "replacement", "********", "Replacement string for custom patterns")
	rootCmd.AddCommand(maskCmd)
}

func runMask(cmd *cobra.Command, args []string) error {
	env := args[0]
	files := args[1:]

	chain, err := buildChainFromFiles(files)
	if err != nil {
		return fmt.Errorf("building chain: %w", err)
	}

	opts := config.MaskOptions{
		MaskSecrets: maskSecrets,
		Rules:       config.DefaultMaskRules(),
	}
	for _, p := range maskPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		opts.Rules = append(opts.Rules, config.MaskRule{
			Pattern:     re,
			Replacement: maskReplacement,
		})
	}

	entries, err := config.MaskLayer(chain, env, opts)
	if err != nil {
		return fmt.Errorf("masking layer: %w", err)
	}

	switch maskFormat {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(entries)
	default:
		changed := 0
		for _, e := range entries {
			marker := " "
			if e.Changed {
				marker = "~"
				changed++
			}
			fmt.Fprintf(os.Stdout, "%s %s=%s\n", marker, e.Key, e.Masked)
		}
		if changed > 0 {
			fmt.Fprintf(os.Stderr, "%d value(s) masked.\n", changed)
		}
		return nil
	}
}
