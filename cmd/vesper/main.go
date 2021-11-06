package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/robotii/vesper"
)

func main() {
	var help, version, compile, verbose, debug bool
	var path string
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&version, "version", false, "shows the current version")
	flag.BoolVar(&compile, "compile", false, "compile the file and output code")
	flag.BoolVar(&verbose, "verbose", false, "verbose mode, print extra information")
	flag.BoolVar(&debug, "debug", false, "debug mode, print extra information about compilation")
	flag.StringVar(&path, "path", "", "add directories to vesper load path")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(1)
	}
	if version {
		fmt.Println(vesper.Version)
		os.Exit(0)
	}
	args := flag.Args()
	interactive := len(args) == 0
	vesper.Init()

	// We create and initialise a new VM here,
	// so that we have a clean environment for the interactive mode
	vm := vesper.NewVM().Init()
	if path != "" {
		for _, p := range strings.Split(path, ":") {
			expandedPath := vesper.ExpandFilePath(p)
			if vesper.IsDirectoryReadable(expandedPath) {
				vm.AddVesperDirectory(expandedPath)
				if debug {
					vesper.Println("[added directory to path: '", expandedPath, "']")
				}
			} else if debug {
				vesper.Println("[directory not readable, cannot add to path: '", expandedPath, "']")
			}
		}
	}
	if !interactive {
		if compile {
			for _, filename := range args {
				generated, err := vm.CompileFile(filename)
				if err != nil {
					vesper.Fatal("*** ", err)
				}
				vesper.Println(generated)
			}
		} else {
			vm.SetFlags(verbose, debug, interactive)
			vm.Run(args...)
		}
	} else {
		vm.SetFlags(verbose, debug, interactive)
		vesper.REPL(vm)
	}
}
