package cmptest_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/spf13/cobra"

	"github.com/coxley/complete"
	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmpcobra"
	"github.com/coxley/complete/cmptest"
	"github.com/coxley/complete/predict"
)

func TestBasic(t *testing.T) {
	cp := complete.NopParser(complete.Command{
		Sub: nil,
		Flags: complete.Flags{
			"--force":            predict.Nothing,
			"--help":             predict.Nothing,
			"--version":          predict.Nothing,
			"--target-directory": predict.Dirs("*"),
		},
		// This is the default when there are no sub-commands set
		Args: predict.Files("*"),
	})
	cmptest.Assert(t, cp, "root --", slices.Collect(maps.Keys(cp.Command().Flags)))
}

func TestCustomPredictor(t *testing.T) {
	cp := complete.NopParser(complete.Command{
		Sub: nil,
		Flags: complete.Flags{
			"--names": predict.Func(func(args.Args) []string {
				return []string{"coxley", "posener"}
			}),
		},
	})
	cmptest.Assert(t, cp, "root --names <TAB>", []string{"coxley", "posener"})
}

func TestCobra(t *testing.T) {
	cmd := &cobra.Command{
		Use:       "count",
		ValidArgs: []string{"one", "two", "three"},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cmptest.Assert(t, cmpcobra.New(cmd), "count <TAB>", []string{"one", "two", "three"})
}

func Example() {
	// Example functions can't actually run tests themselves so leaving this empty.
	//
	// Just use the testcases as examples
	_ = TestBasic
	_ = TestCustomPredictor
	_ = TestCobra
}
