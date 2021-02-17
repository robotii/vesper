package vesper

import (
	"flag"
	"fmt"
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
// Version - this version of vesper
const Version = "vesper v0.1"

// Main entrypoint for main Vesper interpreter
func Main(extns ...Extension) {
	var path string
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&version, "version", false, "Show the current version")
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
