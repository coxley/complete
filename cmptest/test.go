package cmptest

import (
	"bufio"
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/coxley/complete"
	"github.com/coxley/complete/internal"
)

// TabMarker can be placed in the prompt to show where the hypothetical user pressed
// TAB
const TabMarker = "<TAB>"

// Assert that the parser returns the correct suggestions given the prompt
//
// See the docs for [Suggestions] on how the prompt should look
func Assert(t *testing.T, cp complete.CommandParser, prompt string, want []string) {
	t.Helper()
	t.Run(prompt, func(t *testing.T) {
		got := Suggestions(t, cp, prompt)
		if len(got) != len(want) {
			t.Fatalf("suggestions don't match, want=%v got=%v", want, got)
		}

		slices.Sort(want)
		slices.Sort(got)

		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("suggestions don't match, want=%v got=%v", want, got)
			}
		}
	})
}

// Suggestions returns the options returned by the parser with a given prompt
//
// The prompt should look like it would on the command-line. If '<TAB>' is included,
// that is where we will assume the user pressed the tab key. The end of the prompt is
// used otherwise.
//
// Example: "mycli sub --<TAB> --other"
func Suggestions(t testing.TB, cp complete.CommandParser, prompt string) []string {
	t.Helper()
	internal.SetupLogging()

	// Determine where the cursor is, assuming end of the prompt if tab marker is
	// missing.
	var compLine string
	var compPoint int
	if ti := strings.Index(prompt, TabMarker); ti != -1 {
		prompt = prompt[:ti] + prompt[ti+len(TabMarker):]
		compLine = prompt
		compPoint = ti
	} else {
		compLine = prompt
		compPoint = len(prompt)
	}

	t.Setenv("COMP_LINE", compLine)
	t.Setenv("COMP_POINT", fmt.Sprint(compPoint))

	// For debugging, point an arrow where the TAB occured
	t.Logf("COMP_LINE: %q", compLine)
	pointed := strings.Repeat(" ", compPoint) + "^"
	t.Logf("COMP_LINE:  %s", pointed)
	t.Logf("COMP_POINT: %d", compPoint)

	// What would be written to the screen gets written here instead
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	// 'name' is only used for the install CLI
	ok := complete.New2F(w, cp).Complete()
	if !ok {
		t.Fatal("expected completion to run")
	}

	err := w.Flush()
	if err != nil {
		t.Fatalf("flushing to buffer: %v", err)
	}

	suggestions := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(suggestions) == 1 && suggestions[0] == "" {
		return nil
	}
	return suggestions
}
