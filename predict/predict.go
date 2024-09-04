package predict

import (
	"github.com/coxley/complete/args"
)

// Predictor implements a predict method, in which given
// command line arguments returns a list of options it predicts.
//
// It's given context about what the user has typed for the predictor to make
// the decision.
type Predictor interface {
	Predict(args.Args) []string
}

// Or unions two predicate functions, so that the result predicate
// returns the union of their predication
func Or(predictors ...Predictor) Predictor {
	return Func(func(a args.Args) (prediction []string) {
		for _, p := range predictors {
			if p == nil {
				continue
			}
			prediction = append(prediction, p.Predict(a)...)
		}
		return
	})
}

// Func determines what terms can follow a command or a flag
// It is used for auto completion, given last - the last word in the already
// in the command line, what words can complete it.
func Func(inner func(args.Args) []string) Predictor {
	return &wrapped{inner}
}

type wrapped struct {
	inner func(args.Args) []string
}

// Predict invokes the predict function and implements the Predictor interface
func (w *wrapped) Predict(a args.Args) []string {
	if w.inner == nil {
		return nil
	}
	return w.inner(a)
}

// Nothing does not expect anything after.
var Nothing Predictor

// Anything expects something, but nothing particular, such as a number
// or arbitrary name.
var Anything = Func(func(args.Args) []string { return nil })
