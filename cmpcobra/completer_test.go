package cmpcobra

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmptest"
	"github.com/coxley/complete/predict"
)

func TestQuery(t *testing.T) {
	cmd := &cobra.Command{
		Use:       "query",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"table1", "table2"},
	}
	cmd.Flags().StringSliceP("column", "c", nil, "column to select")

	RegisterFlag(cmd, "column", predict.Func(func(args args.Args) []string {
		cmd := args.ParsedRoot.(*cobra.Command)
		posArgs := cmd.Flags().Args()

		if len(posArgs) == 0 {
			return nil
		}

		table := posArgs[0]
		switch table {
		case "table1":
			return []string{"colA", "colB", "colC"}
		case "table2":
			return []string{"colX", "colY", "colZ"}
		default:
			return nil
		}
	}))

	cmptest.Assert(t, New(cmd), "query table1 -c <TAB>", []string{"colA", "colB", "colC"})
	cmptest.Assert(t, New(cmd), "query table2 -c <TAB>", []string{"colX", "colY", "colZ"})
	cmptest.Assert(t, New(cmd), "query table2 -c <TAB> -c colX", []string{"colX", "colY", "colZ"})
	cmptest.Assert(t, New(cmd), "query table3 -c <TAB>", []string{})
	cmptest.Assert(t, New(cmd), "query -c <TAB>", []string{})
	// Because we don't currently support forward-referencing past the TAB
	cmptest.Assert(t, New(cmd), "query -c <TAB> -c colX table2", []string{})
}

func TestPersistentRegister(t *testing.T) {
	root := &cobra.Command{
		Use: "root",
	}
	root.PersistentFlags().String("env", "", "")
	pred := predict.Func(func(args args.Args) []string {
		return []string{"foo"}
	})
	RegisterFlag(root, "env", pred)

	cmptest.Assert(t, New(root), "root --env <TAB>", []string{"foo"})
}

func TestPersistent(t *testing.T) {
	// Test that the child command can get access to persistent flags without looking
	// at the root
	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{
			name:   "none set",
			prompt: "root child ",
			want:   "",
		},
		{
			name:   "root set",
			prompt: "root --env prod child ",
			want:   "prod",
		},
		{
			name:   "child set",
			prompt: "root child --env prod ",
			want:   "prod",
		},
	}
	root := &cobra.Command{
		Use: "root",
	}
	root.PersistentFlags().String("env", "", "")

	child := &cobra.Command{
		Use: "child",
	}
	root.AddCommand(child)

	for _, tt := range tests {
		pred := predict.Func(func(args args.Args) []string {
			root := args.ParsedRoot.(*cobra.Command)
			child := root.Commands()[0]
			got, err := child.Flags().GetString("env")
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			return nil
		})
		RegisterCmd(child, pred)
		cmptest.Assert(t, New(root), tt.prompt, []string{})
	}
}

func TestHiddenFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "root"}
	cmd.Flags().String("hidden", "", "deprecated flag")
	err := cmd.Flags().MarkHidden("hidden")
	require.NoError(t, err)

	cmptest.Assert(t, New(cmd), "root -<TAB>", []string{})
	cmptest.Assert(t, New(cmd, ShowHiddenFlags(true)), "root -<TAB>", []string{"--hidden"})
}
