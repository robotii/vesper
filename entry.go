package kana

import (
	"flag"
	"os"
)

// Main entrypoint for main Kana interpreter
func Main() {
	var help bool
	var path string
	flag.BoolVar(&help, "help", false, "Show help")
	flag.StringVar(&path, "path", "", "add directories to kana load path")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(1)
	}
	args := flag.Args()

	if len(args) > 0 {
		Run(args...)
	} else {
		REPL()
	}

	Cleanup()
}

// Run executes the requested files
func Run(args ...string) {
}

// REPL starts a REPL
func REPL() {
}

// Cleanup deinitialise the runtime
func Cleanup() {
}
