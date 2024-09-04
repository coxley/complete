package main

import (
	"testing"

	"github.com/coxley/complete/cmpcobra"
	"github.com/coxley/complete/cmptest"
)

func TestIncrementing(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{
			name:   "root",
			prompt: "incrementing --num <TAB>",
			want:   []string{"1"},
		},
		{
			name:   "child",
			prompt: "incrementing --num 1 child --num <TAB>",
			want:   []string{"2"},
		},
		{
			name:   "sub-child",
			prompt: "incrementing --num 1 child --num 2 sub-child --num <TAB>",
			want:   []string{"3"},
		},
		{
			name:   "child only",
			prompt: "incrementing child --num <TAB>",
			want:   []string{"1"},
		},
		{
			name:   "sub-child only",
			prompt: "incrementing child sub-child --num <TAB>",
			want:   []string{"1"},
		},
		// TODO: For right now, backtracked tabs only have context from the TAB and before.
		// The completion func would not be able to make decisions on sub-child's value
		// here
		{
			name:   "child backtracking",
			prompt: "incrementing --num 1 child --num <TAB> sub-child --num 3",
			want:   []string{"2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completer := cmpcobra.New(RootCommand())
			cmptest.Assert(t, completer, tt.prompt, tt.want)
		})
	}
}
