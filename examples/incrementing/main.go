package main

import (
	"cmp"
	"fmt"
	"os"
	"strconv"

	"github.com/coxley/complete"
	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmpcobra"
	"github.com/coxley/complete/predict"
	"github.com/spf13/cobra"
)

func main() {
	cmd := RootCommand()
	// We're not using the cobra completions so don't suggest it in help output
	cmd.CompletionOptions.DisableDefaultCmd = true

	// If tab-completion takes place, exit
	if complete.New2(cmpcobra.New(cmd)).Complete() {
		return
	}

	// Otherwise proceed as usual
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "incrementing",
	}
	root.Flags().Int("num", 0, "")
	cmpcobra.RegisterFlag(root, "num", predict.Func(predictNextValue))

	child := ChildCommand()
	child.AddCommand(SubChildCommand())
	root.AddCommand(child)
	return root
}

func ChildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "child",
	}
	cmd.Flags().Int("num", 0, "")
	cmpcobra.RegisterFlag(cmd, "num", predict.Func(predictNextValue))
	return cmd
}

func SubChildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "sub-child",
	}
	cmd.Flags().Int("num", 0, "")
	cmpcobra.RegisterFlag(cmd, "num", predict.Func(predictNextValue))
	return cmd
}

// predictNextValue returns the next value in the sequence by looking at '--num' values
// set in other commands
func predictNextValue(args args.Args) []string {
	root := args.ParsedRoot.(*cobra.Command)

	rootNum := flagInt(root, "num")
	childNum := flagInt(root.Commands()[0], "num")
	subChildNum := flagInt(root.Commands()[0].Commands()[0], "num")

	// Find the max of the first non-zero value, preferring child commands
	lastSetNum := max(cmp.Or(subChildNum, childNum, rootNum))
	if lastSetNum == 0 {
		return []string{"1"}
	}

	return []string{strconv.Itoa(lastSetNum + 1)}
}

// flagInt coerces a flag into an integer if it is set, otherwise returns 0
func flagInt(cmd *cobra.Command, name string) int {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return 0
	}
	if flag.Value == nil {
		return 0
	}

	n, err := strconv.Atoi(flag.Value.String())
	if err != nil {
		return 0
	}
	return n
}
