package predict

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coxley/complete/args"
	"github.com/coxley/complete/internal"
)

func TestPredicate(t *testing.T) {
	t.Parallel()
	internal.Chdir(t)

	tests := []struct {
		name    string
		p       Predictor
		argList []string
		want    []string
	}{
		{
			name: "set",
			p:    Set("a", "b", "c"),
			want: []string{"a", "b", "c"},
		},
		{
			name: "set/empty",
			p:    Set(),
			want: []string{},
		},
		{
			name: "anything",
			p:    Anything,
			want: []string{},
		},
		{
			name: "or: word with nil",
			p:    Or(Set("a"), nil),
			want: []string{"a"},
		},
		{
			name: "or: nil with word",
			p:    Or(nil, Set("a")),
			want: []string{"a"},
		},
		{
			name: "or: nil with nil",
			p:    Or(Nothing, Nothing),
			want: []string{},
		},
		{
			name: "or: word with word with word",
			p:    Or(Set("a"), Set("b"), Set("c")),
			want: []string{"a", "b", "c"},
		},
		{
			name: "files/txt",
			p:    Files("*.txt"),
			want: []string{"./", "dir/", "outer/", "a.txt", "b.txt", "c.txt", ".dot.txt"},
		},
		{
			name:    "files/txt",
			p:       Files("*.txt"),
			argList: []string{"./dir/"},
			want:    []string{"./dir/"},
		},
		{
			name:    "complete files inside dir if it is the only match",
			p:       Files("foo"),
			argList: []string{"./dir/", "./d"},
			want:    []string{"./dir/", "./dir/foo"},
		},
		{
			name:    "complete files inside dir when argList includes file name",
			p:       Files("*"),
			argList: []string{"./dir/f", "./dir/foo"},
			want:    []string{"./dir/foo"},
		},
		{
			name:    "files/md",
			p:       Files("*.md"),
			argList: []string{""},
			want:    []string{"./", "dir/", "outer/", "readme.md"},
		},
		{
			name:    "files/md with ./ prefix",
			p:       Files("*.md"),
			argList: []string{".", "./"},
			want:    []string{"./", "./dir/", "./outer/", "./readme.md"},
		},
		{
			name:    "dirs",
			p:       Dirs("*"),
			argList: []string{"di", "dir", "dir/"},
			want:    []string{"dir/"},
		},
		{
			name:    "dirs with ./ prefix",
			p:       Dirs("*"),
			argList: []string{"./di", "./dir", "./dir/"},
			want:    []string{"./dir/"},
		},
		{
			name:    "predict anything in dir",
			p:       Files("*"),
			argList: []string{"dir", "dir/", "di"},
			want:    []string{"dir/", "dir/foo", "dir/bar"},
		},
		{
			name:    "predict anything in dir with ./ prefix",
			p:       Files("*"),
			argList: []string{"./dir", "./dir/", "./di"},
			want:    []string{"./dir/", "./dir/foo", "./dir/bar"},
		},
		{
			name:    "root directories",
			p:       Dirs("*"),
			argList: []string{""},
			want:    []string{"./", "dir/", "outer/"},
		},
		{
			name:    "root directories with ./ prefix",
			p:       Dirs("*"),
			argList: []string{".", "./"},
			want:    []string{"./", "./dir/", "./outer/"},
		},
		{
			name:    "nested directories",
			p:       Dirs("*.md"),
			argList: []string{"ou", "outer", "outer/"},
			want:    []string{"outer/", "outer/inner/"},
		},
		{
			name:    "nested directories with ./ prefix",
			p:       Dirs("*.md"),
			argList: []string{"./ou", "./outer", "./outer/"},
			want:    []string{"./outer/", "./outer/inner/"},
		},
		{
			name:    "nested inner directory",
			p:       Files("*.md"),
			argList: []string{"outer/i"},
			want:    []string{"outer/inner/", "outer/inner/readme.md"},
		},
	}

	for _, tt := range tests {

		// no args in argList, means an empty argument
		if len(tt.argList) == 0 {
			tt.argList = append(tt.argList, "")
		}

		for _, arg := range tt.argList {
			t.Run(tt.name+"/arg="+arg, func(t *testing.T) {
				matches := tt.p.Predict(args.New(arg, nil))

				sort.Strings(matches)
				sort.Strings(tt.want)

				got := strings.Join(matches, ",")
				want := strings.Join(tt.want, ",")

				if got != want {
					t.Errorf("failed %s\ngot = %s\nwant: %s", t.Name(), got, want)
				}
			})
		}
	}
}

func TestCached(t *testing.T) {
	t.Parallel()
	internal.SetupLogging()

	var hits int
	allServices := []string{"foo", "bar", "baz"}
	pred := Cached("services", time.Second, func() []string {
		hits++
		return allServices
	})

	got := pred.Predict(args.New(" ", nil))
	require.Equal(t, allServices, got)
	require.Equal(t, 1, hits)
	hits = 0

	got = pred.Predict(args.New(" ", nil))
	require.Equal(t, allServices, got)
	require.Equal(t, 0, hits)

	time.Sleep(time.Second + time.Millisecond)
	got = pred.Predict(args.New(" ", nil))
	require.Equal(t, allServices, got)
	require.Equal(t, 1, hits)
}

func TestMatchFile(t *testing.T) {
	t.Parallel()
	internal.Chdir(t)

	type matcherTest struct {
		prefix string
		want   bool
	}

	tests := []struct {
		long  string
		tests []matcherTest
	}{
		{
			long: "file.txt",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "f", want: true},
				{prefix: "./f", want: true},
				{prefix: "./.", want: false},
				{prefix: "file.", want: true},
				{prefix: "./file.", want: true},
				{prefix: "file.txt", want: true},
				{prefix: "./file.txt", want: true},
				{prefix: "other.txt", want: false},
				{prefix: "/other.txt", want: false},
				{prefix: "/file.txt", want: false},
				{prefix: "/fil", want: false},
				{prefix: "/file.txt2", want: false},
				{prefix: "/.", want: false},
			},
		},
		{
			long: "./file.txt",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "f", want: true},
				{prefix: "./f", want: true},
				{prefix: "./.", want: false},
				{prefix: "file.", want: true},
				{prefix: "./file.", want: true},
				{prefix: "file.txt", want: true},
				{prefix: "./file.txt", want: true},
				{prefix: "other.txt", want: false},
				{prefix: "/other.txt", want: false},
				{prefix: "/file.txt", want: false},
				{prefix: "/fil", want: false},
				{prefix: "/file.txt2", want: false},
				{prefix: "/.", want: false},
			},
		},
		{
			long: "/file.txt",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: "f", want: false},
				{prefix: "./f", want: false},
				{prefix: "./.", want: false},
				{prefix: "file.", want: false},
				{prefix: "./file.", want: false},
				{prefix: "file.txt", want: false},
				{prefix: "./file.txt", want: false},
				{prefix: "other.txt", want: false},
				{prefix: "/other.txt", want: false},
				{prefix: "/file.txt", want: true},
				{prefix: "/fil", want: true},
				{prefix: "/file.txt2", want: false},
				{prefix: "/.", want: false},
			},
		},
		{
			long: "./",
			tests: []matcherTest{
				{prefix: "", want: true},
				{prefix: ".", want: true},
				{prefix: "./", want: true},
				{prefix: "./.", want: false},
			},
		},
	}

	for _, tt := range tests {
		for _, ttt := range tt.tests {
			name := fmt.Sprintf("long=%q&prefix=%q", tt.long, ttt.prefix)
			t.Run(name, func(t *testing.T) {
				got := matchFile(tt.long, ttt.prefix)
				if got != ttt.want {
					t.Errorf("Failed %s: got = %t, want: %t", name, got, ttt.want)
				}
			})
		}
	}
}
