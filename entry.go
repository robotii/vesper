package vesper

import (
	"flag"
	"os"
)

// Extension allows for extending Vesper
type Extension interface {
	Init() error
	Cleanup()
	String() string
}

// Version - this version of vesper
const Version = "vesper v0.1"

var verbose bool
var debug bool
var interactive bool

var extensions []Extension
// SetFlags - set various flags controlling the vm
func SetFlags(v bool, d bool, i bool) {
	verbose = v
	debug = d
	interactive = i
}
// Init initialise the base environment and extensions
func Init(extns ...Extension) {
}

// Cleanup deinitialise the extensions
func Cleanup() {
	for _, ext := range extensions {
		ext.Cleanup()
	}
}

// Run the given files in the vesper vm
func Run(args ...string) {
}
	// TODO: implement

// Main entrypoint for main Vesper interpreter
func Main(extns ...Extension) {
	var help, compile, optimize, verbose, debug bool
	var path string
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&compile, "compile", false, "compile the file and output code")
	flag.BoolVar(&optimize, "optimize", false, "optimize execution speed, should work for correct code, relaxes some checks")
	flag.BoolVar(&verbose, "verbose", false, "verbose mode, print extra information")
	flag.BoolVar(&debug, "debug", false, "debug mode, print extra information about compilation")
	flag.StringVar(&path, "path", "", "add directories to vesper load path")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(1)
	}

	if version {
		fmt.Println(Version)
		os.Exit(0)
	}
	args := flag.Args()

	if len(args) > 0 {
		Run(args...)
	} else {
		REPL()
	}

	Cleanup()
}
