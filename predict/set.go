package predict

import "github.com/coxley/complete/args"

// Set expects specific set of terms, given in the options argument.
func Set(options ...string) Predictor {
	return predictSet(options)
}

type predictSet []string

func (p predictSet) Predict(a args.Args) []string {
	return p
}
