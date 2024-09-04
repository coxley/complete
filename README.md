# complete

[![Go Reference](https://pkg.go.dev/badge/github.com/coxley/complete.svg)](https://pkg.go.dev/github.com/coxley/complete)

This package provides a library for writing shell-agnostic tab completion.
(bash, zsh, fish)

While completion for any program can be written with the library, it's targeted for
self-completing Go programs. This is a program that doesn't need to distribute, or
maintain, any complex bash/zsh/fish script alongside their normal CLI.

It was originally forked from [posener/complete](https://github.com/posener/complete),
which has went in a different direction.

## Features

- Describe the skeleton of your CLI and get tab-completion for free
- Write custom predictors to enrich suggestions using your existing business logic
- Dynamic, contextual suggestions based on what the user has typed. 
    - A flag may need different values depending on a positional argument's value.
- Helpers to generate skeletons from well-known CLI frameworks. 
    - `cobra` is automatic
    - Seamless helpers for `flags` and `urfavecli` is planned.

## Self Completion

Shells can query external binaries to get tab suggestions. 

They provide a few environment variables to the program, and this package parses them.
It notes what has been typed in the prompt, the cursor position where the user has
pressed TAB, and returns relevant suggestions.

By writing custom predictors, tools can hook into this and enrich the user's
experience. Use existing logic in your code without duplicating it to a shell script.

## Installation

While developers don't need to distribute complicated shell programs for completion,
a bit of configuration is still needed.

Bash is one command and zsh is two, fish is a bit more. Any program that uses this
package can be installed by setting `COMP_INSTALL=1` and running the program.

```bash
# Detects the user's shell and operating system, configuring them appropriately.
COMP_INSTALL=1 mycli
COMP_INSTALL=1 COMP_YES=1 mycli # To auto confirm

# Uninstall
COMP_UNINSTALL=1 mycli
```

If you prefer the manual way:

```bash
# Bash
# The last argument is what the user is typing as argv[0] - not a path
complete -C /path/to/mycli mycli

# Zsh
autoload -U +X bashcompinit && bashcompinit
complete -C /path/to/mycli mycli
```

# Examples

If you want to jump into an example, here they are:

- Runnable programs: [./examples](./examples)
- Go examples: 


# Predictors

A `Predictor` is any type that implements `Predict(args.Args) []string`.

`args.Args` contains a few fields:

```go
type Args struct {
    // Arguments in typed by the user so far, up until they pressed TAB.
    //   - At some point this will be all arguments, even if TAB was pressed in the
    //     middle of the line.
    All []string
    // Same as above, excluding the one currently being typed.
    Completed []string
    // The word currently being typed, or empty if there's a space before where 
    // TAB was pressed.
    Last string
    // Last fully-typed word
    LastCompleted string
    // Domain-specific value that was emitted by `args.Parser(all []string)`
    ParsedRoot any
}
```

Each `Predictor` is mapped to a flag or command to generate suggestions depending on
where the user presses TAB. If no predictor is set for a command, it's sub-commands are
used. Otherwise it defaults to `predict.Files`.

There are a few canonical predictors to help you get started:

```go
predict.Anything
predict.Cached
predict.Dirs
predict.Files
predict.Func
predict.Nothing
predict.Or
predict.ScopedCache
predict.Set
```

# Testing

To make testing easy, the `cmptest` package provides two functions:

```go
// Suggestions returns the options returned by the parser with a given prompt
//
// The prompt should look like it would on the command-line. If '<TAB>' is included,
// that is where we will assume the user pressed the tab key. The end of the prompt is
// used otherwise.
//
// Example: "mycli sub --<TAB> --other"
func Suggestions(t testing.TB, cp complete.CommandParser, prompt string) []string

// Assert suggestions from [Suggestions]
func Assert(t *testing.T, cp complete.CommandParser, prompt string, want []string)
```

A basic example using `cobra` would look like:

```go
func TestBasic(t *testing.T) {
    cmd := &cobra.Command{
        Use:   "count",
        ValidArgs: []string{"one", "two", "three"},
        CompletionOptions: cobra.CompletionOptions{
            DisableDefaultCmd: true,
        },
    }

    cmptest.Assert(t, cmpcobra.New(cmd), "count <TAB>", []string{"one", "two", "three"})
}
```

# Fork Differences

First, much thanks and credit to posener/complete. It was the first library that
demonstrated self-completion to me many years ago.

I've been using it ever sincce until recently, mostly due to it's `v2` version making
decisions that make dynamic, contextual completion hard.

- Assumptions around certain flags always existing (`--help`, and `-h`)
    - While I agree every CLI should define these, it's not the completion engine's
      place to assert.

- Predictors only having access to their token
    - Tab completion should enrich a user's experience, and sometimes this means making
      different decisions depending on other parts of the prompt.


As such, this is a fork from `v1`. I've had to change a few core ways in how the
library works, and decided to publish the results going forward. You should be able to
use this library by simply changing the import - all of the exported symbols have been
aliased.

## Changes

- Removal of `cmd install` in favor of v2's `COMP_INSTALL` semantics
- `New2` and `New2F` functions as primary entrypoints
- The concept of a `Parser` that can give wider context to each `Predictor`
- The concept of a `Commander` that can return a `Command`
    - Enables framework-aware helpers to generate a completion skeleton
- New packages `args` and `predict` for scoping and import cycle issues
- `args.Args.ParsedRoot` contains the result of `Parser.Parse()`
- `predict.Cached` to re-use values that may need a network or expensive call to
   generate
- New packages `cmptest` and `cmplog` for easier testing
- New package `cmpcobra` for generating completion from `cobra` programs
