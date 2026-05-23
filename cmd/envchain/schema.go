package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/user/envchain/internal/config"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Validate resolved env values against a JSON schema file",
	RunE:  runSchema,
}

var (
	schemaFile string
	schemaEnv  string
)

func init() {
	schemaCmd.Flags().StringVar(&schemaFile, "schema", "", "path to JSON schema file (required)")
	schemaCmd.Flags().StringVar(&schemaEnv, "env", "dev", "environment to resolve before validation")
	_ = schemaCmd.MarkFlagRequired("schema")
	rootCmd.AddCommand(schemaCmd)
}

// schemaFileDef is the on-disk representation of a schema.
type schemaFileDef struct {
	Fields []struct {
		Key      string `json:"key"`
		Required bool   `json:"required"`
		Pattern  string `json:"pattern"`
		Secret   bool   `json:"secret"`
	} `json:"fields"`
}

func runSchema(cmd *cobra.Command, args []string) error {
	raw, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	var def schemaFileDef
	if err := json.Unmarshal(raw, &def); err != nil {
		return fmt.Errorf("parsing schema file: %w", err)
	}

	fields := make([]config.SchemaField, 0, len(def.Fields))
	for _, f := range def.Fields {
		fields = append(fields, config.SchemaField{
			Key:      f.Key,
			Required: f.Required,
			Pattern:  f.Pattern,
			Secret:   f.Secret,
		})
	}

	schema, err := config.NewSchema(fields)
	if err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	if len(args) == 0 {
		return fmt.Errorf("at least one config file must be provided")
	}

	chain, err := buildChainFromFiles(args)
	if err != nil {
		return err
	}

	resolved, err := chain.Resolve(schemaEnv)
	if err != nil {
		return fmt.Errorf("resolving env %q: %w", schemaEnv, err)
	}

	values := make(map[string]string, len(resolved))
	for k, v := range resolved {
		values[k] = v.Value
	}

	violations := config.ValidateAgainstSchema(schema, values)
	if len(violations) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "schema validation passed for env %q\n", schemaEnv)
		return nil
	}

	for _, v := range violations {
		fmt.Fprintf(cmd.ErrOrStderr(), "FAIL  %s\n", v.Error())
	}
	return fmt.Errorf("schema validation failed with %d violation(s)", len(violations))
}
