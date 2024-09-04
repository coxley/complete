package complete_test

import (
	"github.com/coxley/complete"
	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmpcobra"
	"github.com/coxley/complete/predict"
	"github.com/spf13/cobra"
)

// Barebones tab completion for a very naive 'cp' program
func Example_barebones() {
	// Barebones 'cp' completion
	comp := complete.New2(complete.NopParser(complete.Command{
		Sub: nil,
		Flags: complete.Flags{
			"--force":            predict.Nothing,
			"--help":             predict.Nothing,
			"--version":          predict.Nothing,
			"--target-directory": predict.Dirs("*"),
		},
		// This is the default when there are no sub-commands set
		Args: predict.Files("*"),
	}))

	if comp.Complete() {
		return
	}
}

// Barebones tab completion for a very naive 'git' program
func Example_barebones_SubCommands() {
	comp := complete.New2(complete.NopParser(complete.Command{
		Sub: complete.Commands{
			"switch": complete.Command{
				Flags: complete.Flags{
					"--quiet": predict.Nothing,
				},
				Args: predict.Func(func(a args.Args) []string {
					return []string{"branch1", "branch2", "master"}
				}),
			},
			"commit": complete.Command{
				Flags: complete.Flags{
					"--message": predict.Nothing,
					"--author": predict.Func(func(a args.Args) []string {
						// Maybe this looks up authors that have previously committed
						// and suggests them
						return nil
					}),
				},
				Args: predict.Func(func(a args.Args) []string {
					// This could return valid pathspecs
					return nil
				}),
			},
		},
		GlobalFlags: complete.Flags{
			"--help":    predict.Nothing,
			"--version": predict.Nothing,
		},
		// Args will default to sub-commands if not set
	}))

	if comp.Complete() {
		return
	}
}

// Cobra commands can be given directly into [cmpcobra.New] to generate a skeleton for
// you.
//
// Individual flags and commmands can have their prediction logic overriden.
func Example_cobra() {
	cmd := &cobra.Command{
		Use:   "toggle-log-levels",
		Short: "Toggle log levels on a remote service for a period of time",
		// These will be included in suggestions
		ValidArgs: []string{"debug", "info", "warn", "error"},
		// We're not using the cobra completions so don't suggest it in help output
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	flags := cmd.Flags()
	flags.StringP("service", "s", "all", "Service to toggle logs on")

	// Override the default prediction value for string flags
	cmpcobra.RegisterFlag(cmd, "service", predict.Func(func(a args.Args) []string {
		return []string{"service1", "service2", "service3"}
	}))

	comp := complete.New2(cmpcobra.New(cmd))
	if comp.Complete() {
		return
	}
}
