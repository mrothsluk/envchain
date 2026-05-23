package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourorg/envchain/internal/config"
)

var (
	cloneSkipSecrets bool
	cloneOnlyKeys    string
	cloneFiles       []string
)

func init() {
	cloneCmd := &cobra.Command{
		Use:   "clone <src-env> <dst-env>",
		Short: "Clone one environment layer into another",
		Args:  cobra.ExactArgs(2),
		RunE:  runClone,
	}
	cloneCmd.Flags().BoolVar(&cloneSkipSecrets, "skip-secrets", false, "omit secret keys from the clone")
	cloneCmd.Flags().StringVar(&cloneOnlyKeys, "only-keys", "", "comma-separated list of keys to clone")
	cloneCmd.Flags().StringArrayVarP(&cloneFiles, "file", "f", nil, "layer config files (repeatable)")
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	srcEnv, dstEnv := args[0], args[1]

	chain, err := buildChainFromFiles(cloneFiles)
	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	var only []string
	if cloneOnlyKeys != "" {
		for _, k := range strings.Split(cloneOnlyKeys, ",") {
			if k = strings.TrimSpace(k); k != "" {
				only = append(only, k)
			}
		}
	}

	opts := config.CloneOptions{
		SkipSecrets: cloneSkipSecrets,
		OnlyKeys:    only,
	}

	n, err := config.CloneLayer(chain, srcEnv, dstEnv, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "cloned %d key(s) from %q to %q\n", n, srcEnv, dstEnv)
	return nil
}
