package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/envchain/internal/config"
)

func init() {
	registerCommand("compare", "compare two environments and show key differences", runCompare)
}

func runCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	redact := fs.Bool("redact", true, "redact secret values in output")
	files := fs.String("files", "", "comma-separated list of env files to load")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return fmt.Errorf("compare requires two environment names, e.g. compare dev prod")
	}
	envA, envB := fs.Arg(0), fs.Arg(1)

	chain, err := buildChainFromFiles(*files)
	if err != nil {
		return err
	}

	res, err := config.CompareEnvs(chain, envA, envB, *redact)
	if err != nil {
		return err
	}

	printCompareResult(res)
	return nil
}

func printCompareResult(res *config.CompareResult) {
	fmt.Fprintf(os.Stdout, "Comparing %q vs %q\n\n", res.EnvA, res.EnvB)

	if keys := res.SortedOnlyInA(); len(keys) > 0 {
		fmt.Fprintf(os.Stdout, "Only in %s:\n", res.EnvA)
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "  - %s=%s\n", k, res.OnlyInA[k])
		}
		fmt.Fprintln(os.Stdout)
	}

	if keys := res.SortedOnlyInB(); len(keys) > 0 {
		fmt.Fprintf(os.Stdout, "Only in %s:\n", res.EnvB)
		for _, k := range keys {
			fmt.Fprintf(os.Stdout, "  + %s=%s\n", k, res.OnlyInB[k])
		}
		fmt.Fprintln(os.Stdout)
	}

	if keys := res.SortedDifferent(); len(keys) > 0 {
		fmt.Fprintln(os.Stdout, "Changed values:")
		for _, k := range keys {
			pair := res.Different[k]
			fmt.Fprintf(os.Stdout, "  ~ %s: %q -> %q\n", k, pair[0], pair[1])
		}
		fmt.Fprintln(os.Stdout)
	}

	if len(res.Identical) > 0 {
		fmt.Fprintf(os.Stdout, "Identical keys (%d): %v\n", len(res.Identical), res.Identical)
	}
}
