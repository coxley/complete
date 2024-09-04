package cmpcobra

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/coxley/complete/cmplog"
	"github.com/coxley/complete/command"
	"github.com/coxley/complete/predict"
)

var (
	flagRegistry = make(map[*pflag.Flag]predict.Predictor, 16)
	cmdRegistry  = make(map[*cobra.Command]predict.Predictor, 16)
)

// RegisterFlag for custom completion logic
//
// The default completion type resolves filenames. This can override that with
// something more fitting.
//
// Ideally this would be scoped to a single [Completer], but it's inconvenient for all
// command "factories" to share that value around. The registered values are unique
// pointers so overwriting won't happen.
func RegisterFlag(cmd *cobra.Command, name string, predictor predict.Predictor) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		panic(fmt.Sprintf("flag %q doesn't exist on %q", name, cmd.Name()))
	}
	flagRegistry[flag] = predictor
}

// RegisterCmd for custom completion logic
//
// The default completion resolves sub-command names (if any), valid args,  or
// filenames. This can override that with something more fitting.
//
// Ideally this would be scoped to a single [Completer], but it's inconvenient for all
// command "factories" to share that value around. The registered values are unique
// pointers so overwriting won't happen.
func RegisterCmd(cmd *cobra.Command, predictor predict.Predictor) {
	if cmd == nil {
		panic("recieved nil pointer for 'cmd'")
	}
	cmdRegistry[cmd] = predictor
}

// Completer is aware of Cobra CLI semantics, and can generate a tab completion
// skeleton for them.
type Completer struct {
	root    *cobra.Command
	options options
}

type options struct {
	showHiddenFlags bool
}

type Option func(*options)

// ShowHiddenFlags controls whether flags explicitly marked as hidden should appear in
// suggestions
func ShowHiddenFlags(show bool) Option {
	return func(o *options) {
		o.showHiddenFlags = show
	}
}

func New(root *cobra.Command, opts ...Option) *Completer {
	var options options
	for _, fn := range opts {
		fn(&options)
	}
	return &Completer{root, options}
}

// Command traversed the [*cobra.Command] to create a completion skeleton, substituting
// registered predictors in their appropriate places
func (c *Completer) Command() command.Command {
	return c.createCompletion(c.root)
}

// Parse the arguments and return the most relevant [*cobra.Command] for predictors to
// inspect
//
// Predictors can run cmd.Root() if they want access to other things.
func (c *Completer) Parse(args []string) any {
	// Ignore errors for unknown during parsing - enriching completion shouldn't error
	old := c.root.FParseErrWhitelist
	defer func() {
		c.root.FParseErrWhitelist = old
	}()
	c.root.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}

	// Parse arguments at each command level, fetching the current most-relevant
	// command in the chain and trimmed args that are relevant for it.
	matched, args, err := c.root.Traverse(args)
	if err != nil {
		cmplog.Log("Error traversing root: %v", err)
		return nil
	}

	matched.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
	err = matched.ParseFlags(args)
	if err != nil {
		cmplog.Log("Likely expected failure from partial flags: %v", err)
	}
	return c.root
}

func cmdPredictor(cmd *cobra.Command) predict.Predictor {
	// Shells (bash/zsh) predict files for programs that don't use completion.
	// Match that behavior. Most of the "win" comes from flag and sub-command
	// names anyway.
	pred := predict.Files("*")
	if cmd.HasSubCommands() {
		// When [command.Command.Sub] has values, we will complete those
		pred = predict.Anything
	}

	if pred, ok := cmdRegistry[cmd]; ok {
		cmplog.Log("Custom predictor for %q", cmd.Name())
		return pred
	}

	// Honor valid args if configured
	if len(cmd.ValidArgs) > 0 {
		// Remove any description that may be included in ValidArgs.
		// A description is following a tab character.
		validArgs := make([]string, 0, len(cmd.ValidArgs))
		for _, v := range cmd.ValidArgs {
			validArgs = append(validArgs, strings.SplitN(v, "\t", 2)[0])
		}
		cmplog.Log("Predicting valid args for %q: %v", cmd.Name(), validArgs)
		pred = predict.Set(validArgs...)
	}

	// TODO: Be aware of pflag annotations for required flag
	// Required flags should be suggested by default, and only if they're not already
	// present in the line.
	return pred
}

// createCompletion walks the Cobra command structure to suggest predictions
func (c *Completer) createCompletion(cmd *cobra.Command) command.Command {
	cmp := command.Command{
		Sub:  command.Commands{},
		Args: cmdPredictor(cmd),
		// While Cobra says cmd.Flags() returns persistent flags, it seems to
		// happen after parsing takes place. We want this ready before then -
		// so walk them separately.
		GlobalFlags: c.flagVisitor(cmd.PersistentFlags()),
		Flags:       c.flagVisitor(cmd.Flags()),
	}

	for _, sub := range cmd.Commands() {
		// NOTE: While we could include c.Aliases, they can inflate the suggestions and
		// don't result in less keystrokes
		cmp.Sub[sub.Name()] = c.createCompletion(sub)
	}
	return cmp
}

// flagVisitor walks the flagset, returning commnad.Flags with appropriate predictors
func (c *Completer) flagVisitor(flags *pflag.FlagSet) command.Flags {
	cmpFlags := command.Flags{}
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden && !c.options.showHiddenFlags {
			return
		}

		var predictor predict.Predictor

		predictor = predict.Files("*")

		// Boolean flags stand on their own - no values expected
		typ := flag.Value.Type()
		if typ == "bool" {
			predictor = predict.Nothing
		}

		if p, ok := flagRegistry[flag]; ok {
			predictor = p
		}

		cmpFlags["--"+flag.Name] = predictor
		if short := flag.Shorthand; short != "" {
			cmpFlags["-"+short] = predictor
		}
	})
	return cmpFlags
}
