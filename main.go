// vcs-torture/main.go
// Copyright 2019 Brian Fitzgerald <neurocline@gmail.com>
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.
//
// vcs-torture is a version control system torture test

package main

import (
	"fmt"
	"log"
	"os"

	"vcs-torture/gsos"
	"vcs-torture/vcs"
)

func main() {
	args := parseArgs()
	args.Run()
}

// ----------------------------------------------------------------------------------------------

func (args *CmdArgs) Run() {

	if args.Create {
		args.mustHaveDest()
		args.mustHaveRepo()
		args.mustHaveVcs()

		repo := vcs.NewRepo(args.Dest, args.Repo, args.Vcs)
		repo.SetVerbose(args.Verbose)
		if !repo.Create() {
			log.Fatalf("Couldn't create repo\n")
		}
	}
}

func (args *CmdArgs) mustHaveDest() {
	if args.Dest == "" {
		log.Fatalf("Specify work area with --dest")
	}
}

func (args *CmdArgs) mustHaveRepo() {
	if args.Repo == "" {
		log.Fatalf("Specify repo name with --repo")
	}
}

func (args *CmdArgs) mustHaveVcs() {
	if args.Vcs == "" {
		log.Fatalf("Specify version control system with --vcs")
	}
}

// ----------------------------------------------------------------------------------------------

type CmdArgs struct {
	// Dest points to our working area
	Dest string

	// Vcs is the name of the version control system to test
	Vcs string

	// Repo is the path to the repo to work on
	Repo string

	// Create will make sure a repo exists
	Create bool

	Help    bool
	Verbose bool
}

func usage(fail int) {
	fmt.Printf("Usage: perf [--vcs=<vcs-name>]\n" +
		       "            [-v|--verbose] [-h|--help]\n")
	os.Exit(fail)
}

// Process command-line
func parseArgs() *CmdArgs {
	p := &CmdArgs{}
	p.parse(os.Args[1:])

	if p.Help {
		usage(0)
	}

	return p
}

func (p *CmdArgs) parse(args []string) {
	for _, arg := range args {

		parsersp := func() bool {
			if len(arg) < 2 || arg[0] != '@' {
				return false
			}
			return p.parseResponse(arg[1:])
		}
		parsearg := func(opt string, val *string) bool {
			optlen := len(opt)
			if len(arg) <= optlen || arg[:optlen] != opt {
				return false
			}
			*val = arg[optlen:]
			return true
		}
		parsebool := func(opt string, val *bool) bool {
			if arg != opt {
				return false
			}
			*val = true
			return true
		}

		if !parsersp() &&
		    !parsearg("--dest=", &p.Dest) &&
		    !parsearg("--repo=", &p.Repo) &&
			!parsebool("--create", &p.Create) &&
		    !parsearg("--vcs=", &p.Vcs) &&
			!parsebool("-v", &p.Verbose) &&
			!parsebool("--verbose", &p.Verbose) &&
			!parsebool("-h", &p.Help) &&
			!parsebool("--help", &p.Help) {
			usage(1)
		}
	}
}

// Given @<response-file>, open file <response-file>
// and parse its lines as if they were command-line arguments,
// one per line
func (p *CmdArgs) parseResponse(responsePath string) bool {
	args, err := gsos.FileReadLines(responsePath)
	if err != nil {
		usage(1)
	}

	p.parse(args)
	return true
}
