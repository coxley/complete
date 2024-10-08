package cmplog

import (
	"io"
	"log"
	"os"
)

// Log is used for debugging purposes
// since complete is running on tab completion, it is nice to
// have logs to the stderr (when writing your own completer)
// to write logs, set the COMP_DEBUG environment variable and
// use complete.Log in the complete program
var Log = getLogger()

const Env = "COMP_DEBUG"

func Reset() {
	Log = getLogger()
}

func getLogger() func(format string, args ...interface{}) {
	logfile := io.Discard
	if os.Getenv(Env) != "" {
		logfile = os.Stderr
	}
	return log.New(logfile, "complete ", log.Flags()).Printf
}
