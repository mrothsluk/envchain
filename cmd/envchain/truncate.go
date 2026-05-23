package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/user/envchain/internal/config"
	"github.com/spf13/cobra"
)

var truncateCmd = &cobra.Command{
	Use:   "truncate",
	Short: "Truncate long env values and report what changed",
	RunE:  runTruncate,
}

func init() {
	truncateCmd.Flags().StringP("env", "e", "dev", "environment to resolve")
	truncateCmd.Flags().IntP("max-len", "m", 80, "maximum value length before truncation")
	truncateCmd.Flags().String("suffix", "...", "suffix appended to truncated values")
	truncateCmd.Flags().Bool("skip-secrets", true, "skip truncation of secret keys (they are redacted)")
	truncateCmd.Flags().StringP("format", "f", "text", "output format: text|json")
	truncateCmd.Flags().StringSliceP("file", "c", nil, "config files to load (required)")
	_ = truncateCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(truncateCmd)
}

func runTruncate(cmd *cobra.Command, _ []string) error {
	env, _ := cmd.Flags().GetString("env")
	maxLen, _ := cmd.Flags().GetInt("max-len")
	suffix, _ := cmd.Flags().GetString("suffix")
	skipSecrets, _ := cmd.Flags().GetBool("skip-secrets")
	format, _ := cmd.Flags().GetString("format")
	files, _ := cmd.Flags().GetStringSlice("file")

	if len(files) == 0 {
		return fmt.Errorf("at least one --file is required")
	}

	c, err := config.NewChain(env, files[0])
	if err != nil {
		return fmt.Errorf("loading chain: %w", err)
	}
	for _, f := range files[1:] {
		if err := c.AddLayer(env, f); err != nil {
			return fmt.Errorf("adding layer %s: %w", f, err)
		}
	}

	opts := config.TruncateOptions{
		MaxLen:      maxLen,
		Suffix:      suffix,
		SkipSecrets: skipSecrets,
	}

	out, results, err := config.TruncateLayer(c, env, opts)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		payload := map[string]interface{}{
			"values":  out,
			"results": results,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	default:
		keys := make([]string, 0, len(out))
		for k := range out {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "%s=%s\n", k, out[k])
		}
		var truncated int
		for _, r := range results {
			if r.Truncated {
				truncated++
				fmt.Fprintf(os.Stderr, "truncated: %s (orig len %d)\n", r.Key, r.OrigLen)
			}
		}
		if truncated > 0 {
			fmt.Fprintf(os.Stderr, "%d value(s) truncated\n", truncated)
		}
	}
	return nil
}
