package complete

import (
	"github.com/coxley/complete/args"
	"github.com/coxley/complete/command"
	"github.com/coxley/complete/predict"
)

var (
	// Deprecated: See [predict.Or]
	PredictOr = predict.Or
	// Deprecated: See [predict.Nothing]
	PredictNothing = predict.Nothing
	// Deprecated: See [predict.Anything]
	PredictAnything = predict.Anything
	// Deprecated: See [predict.Dirs]
	PredictDirs = predict.Dirs
	// Deprecated: See [predict.Files]
	PredictFiles = predict.Files
	// Deprecated: See [predict.Set]
	PredictSet = predict.Set
)

type (
	// Deprecated: See [predict.Predictor]
	Predictor = predict.Predictor
	// Deprecated: See [predict.Func]
	PredictFunc = oldFunc
	// Deprecated: See [args.Parser]
	Parser = args.Parser
	// Deprecated: see [args.Args]
	Args = args.Args

	// Alias to [command.Command] for import ergonomics
	Command = command.Command
	// Alias to [command.Commands] for import ergonomics
	Commands = command.Commands
	// Alias to [command.Flags] for import ergonomics
	Flags = command.Flags
)

// Compatibility with posener/complete v1
//
// We've changed this for better godoc ordering in pkgsite
type oldFunc func(args.Args) []string

// Predict invokes the predict function and implements the Predictor interface
func (f oldFunc) Predict(a args.Args) []string {
	if f == nil {
		return nil
	}
	return f(a)
}

var _ predict.Predictor = (oldFunc)(nil)
