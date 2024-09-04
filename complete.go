package complete

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmplog"
	"github.com/coxley/complete/command"
	"github.com/coxley/complete/internal/install"
)

const (
	envLine  = "COMP_LINE"
	envPoint = "COMP_POINT"
)

var Log = cmplog.Log

// Complete structs define completion for a command with CLI options
type Complete struct {
	Command Command
	Out     io.Writer
	Parser  args.Parser
}

// Commander returns a structured [Command]
type Commander interface {
	Command() command.Command
}

// CommandParser should generate a fully-structured [Command] and parse arguments into
// an object that predictors can use.
type CommandParser interface {
	Commander
	args.Parser
}

// NopParser returns a [CommandParser] that returns nil when parsing.
func NopParser(command command.Command) CommandParser {
	return &nopParser{command}
}

type nopParser struct {
	command Command
}

func (nopParser) Parse([]string) any {
	return nil
}

func (p *nopParser) Command() command.Command {
	return p.command
}

// New creates a new complete command.
//
// 'name' is unused, but is kept for backward-compatibility with posener/complete. It
// used to be used for installation of the completion script, but we prefer using
// os.Args[0] to allow the user to control what they name their binaries.
func New(name string, command command.Command) *Complete {
	return &Complete{
		Command: command,
		Out:     os.Stdout,
	}
}

// New2 returns a completer structured by the [CommandParser]
//
// By accepting an [args.Parser], predictors can gain extra insight to the command
// at large to influence their suggestions. The result of [args.Parser.Parse] is stored in
// [args.Args.ParsedRoot] before any predictors run.
//
// Suggestions are printed to [os.Stdout].
func New2(cp CommandParser) *Complete {
	return New2F(os.Stdout, cp)
}

// New2F returns a completer that writes suggestions to 'w'
func New2F(w io.Writer, cp CommandParser) *Complete {
	return &Complete{
		Command: cp.Command(),
		Out:     w,
		Parser:  cp,
	}
}

// Complete determines if the user needs suggestions, and returns true if so. Programs
// should exit when true.
//
// Environment variables that control our logic:
//
//   - COMP_LINE: prompt of the user
//   - COMP_POINT: cursor position wher tab was pressed
//   - COMP_INSTALL=1: install completion script into the user's shell
//   - COMP_UNINSTALL=1: uninstall completion script from the user's shell
//   - COMP_YES=1: don't prompt when installing or uninstall
func (c *Complete) Complete() bool {
	// Install (or uninstall) completion into the user's shell if requested
	doInstall := os.Getenv("COMP_INSTALL") == "1"
	doUninstall := os.Getenv("COMP_UNINSTALL") == "1"
	autoYes := os.Getenv("COMP_YES") == "1"
	if doInstall || doUninstall {
		install.Run(os.Args[0], doUninstall, autoYes, os.Stdout, os.Stdin)
		return true
	}

	line, point, ok := getEnv()
	if !ok {
		return false
	}

	// TODO: Remove. Ideally, we want the full context of what the shell sent us for
	// optimal enrichment, but we may need framework-specific logic for parsing to get
	// there.
	//
	// As is, we will only pass everything up to the tab even if there's more typed on
	// the line. (eg: cursor moved back)
	if point >= 0 && point < len(line) {
		line = line[:point]
	}

	Log("Completing phrase: %s", line)
	a := args.New(line, c.Parser)
	Log("Completing last field: %s", a.Last)
	options := c.Command.Predict(a)
	Log("Options: %s", options)

	// filter only options that match the last argument
	//
	// TODO: Adjust logic so that predictors can control what is matched vs. always
	// prefix based.
	matches := []string{}
	for _, option := range options {
		if strings.HasPrefix(option, a.Last) {
			matches = append(matches, option)
		}
	}
	Log("Matches: %s", matches)
	c.output(matches)
	return true
}

func getEnv() (line string, point int, ok bool) {
	line = os.Getenv(envLine)
	if line == "" {
		return
	}
	point, err := strconv.Atoi(os.Getenv(envPoint))
	if err != nil {
		// If failed parsing point for some reason, set it to point
		// on the end of the line.
		Log("Failed parsing point %s: %v", os.Getenv(envPoint), err)
		point = len(line)
	}
	return line, point, true
}

func (c *Complete) output(options []string) {
	// stdout of program defines the complete options
	for _, option := range options {
		fmt.Fprintln(c.Out, option)
	}
}
