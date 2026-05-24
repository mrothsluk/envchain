package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/envchain/internal/config"
)

var (
	freezeEnv          string
	freezeOutput       string
	freezeFormat       string
	freezeRedactSecret bool
)

func init() {
	freezeCmd := &cobra.Command{
		Use:   "freeze",
		Short: "Freeze a resolved environment layer to a JSON file",
		RunE:  runFreeze,
	}
	freezeCmd.Flags().StringVarP(&freezeEnv, "env", "e", "base", "environment to freeze")
	freezeCmd.Flags().StringVarP(&freezeOutput, "output", "o", "", "path to write frozen JSON (stdout if empty)")
	freezeCmd.Flags().StringVarP(&freezeFormat, "format", "f", "text", "report format: text or json")
	freezeCmd.Flags().BoolVar(&freezeRedactSecret, "redact-secrets", true, "redact secret values in output")
	rootCmd.AddCommand(freezeCmd)
}

func runFreeze(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("freeze: at least one config file required")
	}

	base := args[0]
	c, err := config.NewChain("base", base)
	if err != nil {
		return fmt.Errorf("freeze: load base: %w", err)
	}
	for i, extra := range args[1:] {
		label := fmt.Sprintf("layer%d", i+1)
		if err := c.AddLayer(label, extra); err != nil {
			return fmt.Errorf("freeze: add layer %s: %w", label, err)
		}
	}

	fl, err := config.FreezeLayer(c, freezeEnv, freezeRedactSecret)
	if err != nil {
		return err
	}

	if freezeOutput != "" {
		if err := config.SaveFreeze(fl, freezeOutput); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "frozen %d keys to %s\n", len(fl.Values), freezeOutput)
		return nil
	}

	return config.WriteFreezeReport(os.Stdout, fl, freezeFormat)
}
