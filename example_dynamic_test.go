package complete_test

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/coxley/complete"
	"github.com/coxley/complete/args"
	"github.com/coxley/complete/cmpcobra"
	"github.com/coxley/complete/cmplog"
	"github.com/coxley/complete/predict"
)

// These services are either "request-oriented" or "message-broker-oriented" and have
// different metadata.
var services = map[string]map[string]string{
	"server1": {
		"grpc_addr": "some.host:50051",
	},
	"server2": {
		"grpc_addr": "other.host:50051",
	},
	"consumer1": {
		"pubsub_topic":        "some_topic",
		"pubsub_subscription": "some_topic/some_subscription",
	},
	"consumer2": {
		"pubsub_topic":        "other_topic",
		"pubsub_subscription": "other_topic/other_subscription",
	},
}

// CLI that returns information about a service, but the information available
// changes depending on the type of service.
//
// Dynamic tab completion can take into account other arguments to influence
// suggestions. In this case, we change the --field recommendations based on the
// service positional argument.
func Example_cobraDynamic() {
	cmd := &cobra.Command{
		Use:       "svc_registry",
		Args:      cobra.ExactArgs(1),
		ValidArgs: slices.Collect(maps.Keys(services)),
		RunE:      run,
	}

	cmd.Flags().StringArrayP("field", "f", nil, "service fields to print")
	cmpcobra.RegisterFlag(cmd, "field", predict.Func(predictFields))

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

func run(cmd *cobra.Command, args []string) error {
	// Print each requested field from the service registry
	svc := args[0]
	fields := services[svc]
	wanted, err := cmd.Flags().GetStringArray("field")
	if err != nil {
		return err
	}

	for _, f := range wanted {
		fmt.Printf("%s:\t%q\n", f, fields[f])
	}
	return nil
}

// predictFields returns the available fields for a given service if it's been
// specified on the command-line
func predictFields(args args.Args) []string {
	root, ok := args.ParsedRoot.(*cobra.Command)
	if !ok {
		cmplog.Log("root cobra command not parsed")
	}

	posArgs := root.Flags().Args()
	if len(posArgs) == 0 {
		// No suggestions to give if a service hasn't been specified
		return nil
	}

	validFields := services[posArgs[0]]
	return slices.Collect(maps.Keys(validFields))
}
