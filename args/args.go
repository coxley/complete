package args

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Parser accepts all completed arguments from the command-line and returns
// a domain-specific object representing the root command
//
// Predictors may use this to gain insight into what else has been provided at any
// layer.
type Parser interface {
	Parse(args []string) any
}

// Args describes command line arguments
type Args struct {
	// All lists of all arguments in command line (not including the command itself)
	All []string
	// Completed lists of all completed arguments in command line,
	// If the last one is still being typed - no space after it,
	// it won't appear in this list of arguments.
	Completed []string
	// Last argument in command line, the one being typed, if the last
	// character in the command line is a space, this argument will be empty,
	// otherwise this would be the last word.
	Last string
	// LastCompleted is the last argument that was fully typed.
	// If the last character in the command line is space, this would be the
	// last word, otherwise, it would be the word before that.
	LastCompleted string

	// ParsedRoot is the return value of [Parser.Parse], and should be the root command
	// structure for your CLI framework.
	//
	// It's useful for a more complex, dynamic [Predictor]. For example, returning
	// different options depending on another flag value or positional argument.
	//
	// Always 'nil' when no Parser is provided.
	ParsedRoot any
}

// Directory gives the directory of the current written
// last argument if it represents a file name being written.
// in case that it is not, we fall back to the current directory.
//
// Deprecated.
func (a Args) Directory() string {
	if info, err := os.Stat(a.Last); err == nil && info.IsDir() {
		return fixPathForm(a.Last, a.Last)
	}
	dir := filepath.Dir(a.Last)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "./"
	}
	return fixPathForm(a.Last, dir)
}

func New(line string, parser Parser) Args {
	var (
		all       []string
		completed []string
	)
	parts := splitFields(line)
	if len(parts) > 0 {
		all = parts[1:]
		completed = removeLast(parts[1:])
	}

	var root any
	if parser != nil {
		root = parser.Parse(completed)
	}
	return Args{
		All:           all,
		Completed:     completed,
		Last:          last(parts),
		LastCompleted: last(completed),
		ParsedRoot:    root,
	}
}

// splitFields returns a list of fields from the given command line.
// If the last character is space, it appends an empty field in the end
// indicating that the field before it was completed.
// If the last field is of the form "a=b", it splits it to two fields: "a", "b",
// So it can be completed.
func splitFields(line string) []string {
	parts := strings.Fields(line)

	// Add empty field if the last field was completed.
	if len(line) > 0 && unicode.IsSpace(rune(line[len(line)-1])) {
		parts = append(parts, "")
	}

	// Treat the last field if it is of the form "a=b"
	parts = splitLastEqual(parts)
	return parts
}

func splitLastEqual(line []string) []string {
	if len(line) == 0 {
		return line
	}
	parts := strings.Split(line[len(line)-1], "=")
	return append(line[:len(line)-1], parts...)
}

// From returns a copy of Args of all arguments after the i'th argument.
func (a Args) From(i int) Args {
	if i >= len(a.All) {
		i = len(a.All) - 1
	}
	a.All = a.All[i+1:]

	if i >= len(a.Completed) {
		i = len(a.Completed) - 1
	}
	a.Completed = a.Completed[i+1:]
	return a
}

func removeLast(a []string) []string {
	if len(a) > 0 {
		return a[:len(a)-1]
	}
	return a
}

func last(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[len(args)-1]
}

// fixPathForm changes a file name to a relative name
func fixPathForm(last string, file string) string {
	// get wording directory for relative name
	workDir, err := os.Getwd()
	if err != nil {
		return file
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		return file
	}

	// if last is absolute, return path as absolute
	if filepath.IsAbs(last) {
		return fixDirPath(abs)
	}

	rel, err := filepath.Rel(workDir, abs)
	if err != nil {
		return file
	}

	// fix ./ prefix of path
	if rel != "." && strings.HasPrefix(last, ".") {
		rel = "./" + rel
	}

	return fixDirPath(rel)
}

func fixDirPath(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
