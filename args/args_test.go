package args

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coxley/complete/internal"
)

func TestArgs(t *testing.T) {
	internal.Chdir(t)

	t.Parallel()
	tests := []struct {
		line          string
		completed     string
		last          string
		lastCompleted string
	}{
		{
			line:          "a b c",
			completed:     "b",
			last:          "c",
			lastCompleted: "b",
		},
		{
			line:          "a b ",
			completed:     "b",
			last:          "",
			lastCompleted: "b",
		},
		{
			line:          "",
			completed:     "",
			last:          "",
			lastCompleted: "",
		},
		{
			line:          "a",
			completed:     "",
			last:          "a",
			lastCompleted: "",
		},
		{
			line:          "a ",
			completed:     "",
			last:          "",
			lastCompleted: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			a := New(tt.line, nil)

			if got, want := strings.Join(a.Completed, " "), tt.completed; got != want {
				t.Errorf("%s failed: Completed = %q, want %q", t.Name(), got, want)
			}
			if got, want := a.Last, tt.last; got != want {
				t.Errorf("Last = %q, want %q", got, want)
			}
			if got, want := a.LastCompleted, tt.lastCompleted; got != want {
				t.Errorf("%s failed: LastCompleted = %q, want %q", t.Name(), got, want)
			}
		})
	}
}

func TestArgs_From(t *testing.T) {
	t.Parallel()
	tests := []struct {
		line         string
		from         int
		newLine      string
		newCompleted string
	}{
		{
			line:         "a b c",
			from:         0,
			newLine:      "b c",
			newCompleted: "b",
		},
		{
			line:         "a b c",
			from:         1,
			newLine:      "c",
			newCompleted: "",
		},
		{
			line:         "a b c",
			from:         2,
			newLine:      "",
			newCompleted: "",
		},
		{
			line:         "a b c",
			from:         3,
			newLine:      "",
			newCompleted: "",
		},
		{
			line:         "a b c ",
			from:         0,
			newLine:      "b c ",
			newCompleted: "b c",
		},
		{
			line:         "a b c ",
			from:         1,
			newLine:      "c ",
			newCompleted: "c",
		},
		{
			line:         "a b c ",
			from:         2,
			newLine:      "",
			newCompleted: "",
		},
		{
			line:         "",
			from:         0,
			newLine:      "",
			newCompleted: "",
		},
		{
			line:         "",
			from:         1,
			newLine:      "",
			newCompleted: "",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%d", tt.line, tt.from), func(t *testing.T) {
			a := New("cmd "+tt.line, nil)
			n := a.From(tt.from)

			assert.Equal(t, tt.newLine, strings.Join(n.All, " "))
			assert.Equal(t, tt.newCompleted, strings.Join(n.Completed, " "))
		})
	}
}

func TestArgs_Directory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		line      string
		directory string
	}{
		{
			line:      "a b c",
			directory: "./",
		},
		{
			line:      "a b c /tm",
			directory: "/",
		},
		{
			line:      "a b c /tmp",
			directory: "/tmp/",
		},
		{
			line:      "a b c /tmp ",
			directory: "./",
		},
		{
			line:      "a b c ./",
			directory: "./",
		},
		{
			line:      "a b c ./dir",
			directory: "./dir/",
		},
		{
			line:      "a b c dir",
			directory: "dir/",
		},
		{
			line:      "a b c ./di",
			directory: "./",
		},
		{
			line:      "a b c ./dir ",
			directory: "./",
		},
		{
			line:      "a b c ./di",
			directory: "./",
		},
		{
			line:      "a b c ./a.txt",
			directory: "./",
		},
		{
			line:      "a b c ./a.txt/x",
			directory: "./",
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			a := New(tt.line, nil)

			if got, want := a.Directory(), tt.directory; got != want {
				t.Errorf("%s failed: directory = %q, want %q", t.Name(), got, want)
			}
		})
	}
}
