package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/envchain/internal/config"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Tag keys in an environment layer",
}

var tagApplyCmd = &cobra.Command{
	Use:   "apply <env> <file> <tag>[,tag...]",
	Short: "Apply tags to all keys in the given environment",
	Args:  cobra.ExactArgs(3),
	RunE:  runTagApply,
}

var tagQueryCmd = &cobra.Command{
	Use:   "query <env> <file> <tag>",
	Short: "List keys in the environment that carry a given tag",
	Args:  cobra.ExactArgs(3),
	RunE:  runTagQuery,
}

var tagFormat string

func init() {
	tagApplyCmd.Flags().StringVar(&tagFormat, "format", "text", "Output format: text|json")
	tagQueryCmd.Flags().StringVar(&tagFormat, "format", "text", "Output format: text|json")
	tagCmd.AddCommand(tagApplyCmd, tagQueryCmd)
	rootCmd.AddCommand(tagCmd)
}

func runTagApply(cmd *cobra.Command, args []string) error {
	env, file, rawTags := args[0], args[1], args[2]
	tags := strings.Split(rawTags, ",")

	c, err := config.NewChain(env, file)
	if err != nil {
		return fmt.Errorf("loading chain: %w", err)
	}

	idx, err := config.TagLayer(c, env, tags)
	if err != nil {
		return err
	}

	if tagFormat == "json" {
		return json.NewEncoder(os.Stdout).Encode(idx)
	}

	for key, keyTags := range idx[env] {
		fmt.Fprintf(os.Stdout, "%s\t[%s]\n", key, strings.Join(keyTags, ", "))
	}
	return nil
}

func runTagQuery(cmd *cobra.Command, args []string) error {
	env, file, tag := args[0], args[1], args[2]

	c, err := config.NewChain(env, file)
	if err != nil {
		return fmt.Errorf("loading chain: %w", err)
	}

	// Tag every key so we can query.
	idx, err := config.TagLayer(c, env, []string{tag})
	if err != nil {
		return err
	}

	entries := config.QueryByTag(idx, tag)

	if tagFormat == "json" {
		return json.NewEncoder(os.Stdout).Encode(entries)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stdout, "no keys found")
		return nil
	}
	for _, e := range entries {
		fmt.Fprintf(os.Stdout, "%s/%s\t[%s]\n", e.Env, e.Key, strings.Join(e.Tags, ", "))
	}
	return nil
}
