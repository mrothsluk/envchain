package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nicholasgasior/envchain/internal/config"
)

func init() {
	var (
		env           string
		files         []string
		removeBlank   bool
		removeSecrets bool
		removeKeys    []string
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove keys from an environment layer",
		Long: `Prune removes keys from the specified environment layer.
Use --blank to drop empty-valued keys, --secrets to drop secret keys,
or --key to target specific keys. Use --dry-run to preview changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrune(env, files, removeBlank, removeSecrets, removeKeys, dryRun)
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "base", "environment layer to prune")
	cmd.Flags().StringArrayVarP(&files, "file", "f", nil, "config files to load (repeatable)")
	cmd.Flags().BoolVar(&removeBlank, "blank", false, "remove keys with empty values")
	cmd.Flags().BoolVar(&removeSecrets, "secrets", false, "remove keys marked as secrets")
	cmd.Flags().StringArrayVarP(&removeKeys, "key", "k", nil, "explicit key to remove (repeatable)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview removals without mutating")

	_ = cmd.MarkFlagRequired("file")

	rootCmd.AddCommand(cmd)
}

func runPrune(env string, files []string, removeBlank, removeSecrets bool, removeKeys []string, dryRun bool) error {
	chain, err := buildChainFromFiles(files)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	opts := config.PruneOptions{
		DryRun:        dryRun,
		RemoveBlank:   removeBlank,
		RemoveSecrets: removeSecrets,
		RemoveKeys:    removeKeys,
	}

	result, err := config.PruneLayer(chain, env, opts)
	if err != nil {
		return err
	}

	if len(result.Removed) == 0 {
		fmt.Fprintln(os.Stdout, "no keys matched prune criteria")
		return nil
	}

	prefix := "removed"
	if dryRun {
		prefix = "would remove"
	}
	fmt.Fprintf(os.Stdout, "%s %d key(s) from [%s]: %s\n",
		prefix, len(result.Removed), env, strings.Join(result.Removed, ", "))
	return nil
}
