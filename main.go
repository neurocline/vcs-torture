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

// main parses all the command-line arguments, running in
// groups, one group for each distinct command
func main() {
	cmd := &Command{args: os.Args[1:]}
	for len(cmd.args) > 0 {
		cmd.Op = ""
		cmd.args = cmd.parse()
		cmd.Run()
	}
}

// ----------------------------------------------------------------------------------------------

func (cmd *Command) Run() {

	if cmd.Op == "create" {
		cmd.mustHaveDest()
		cmd.mustHaveRepo()
		cmd.mustHaveVcs()

		repo := vcs.NewRepo(cmd.Dest, cmd.Repo, cmd.Vcs)
		repo.SetVerbose(cmd.Verbose)
		if !repo.Create() {
			log.Fatalf("Couldn't create repo\n")
		}
	} else if cmd.Op == "remove" {
		cmd.mustHaveDest()
		cmd.mustHaveRepo()
		if !vcs.DeleteRepo(cmd.Dest, cmd.Repo, cmd.Vcs) {
			log.Fatalf("Couldn't remove repo\n")
		}
	} else {
		log.Fatalf("Unknown op: %s\n", cmd.Op)
	}
}

func (cmd *Command) mustHaveDest() {
	if cmd.Dest == "" {
		log.Fatalf("Specify work area with --dest")
	}
}

func (cmd *Command) mustHaveRepo() {
	if cmd.Repo == "" {
		log.Fatalf("Specify repo name with --repo")
	}
}

func (cmd *Command) mustHaveVcs() {
	if cmd.Vcs == "" {
		log.Fatalf("Specify version control system with --vcs")
	}
}

// ----------------------------------------------------------------------------------------------

type Command struct {
	// Op is what we are doing (create, add, ...)
	Op string

	// Dest points to our working area
	Dest string

	// Vcs is the name of the version control system to test
	Vcs string

	// Repo is the path to the repo to work on
	Repo string

	Help    bool
	Verbose bool

	// remaining command-line arguments
	args []string
	print bool
}

func usage(fail int) {
	fmt.Printf("Usage: vcs=torture [--vcs=<vcs-name>]\n" +
		       "            [-v|--verbose] [-h|--help]\n")
	os.Exit(fail)
}

const parseDiagnostic = false

func (cmd *Command) parse() []string {
	if parseDiagnostic && cmd.Verbose {
		fmt.Printf("enter cmd.args=%s\n", cmd.args)
	}

	// since some operations can change the arg list,
	// iterate through by hand
	i := 0
	for i < len(cmd.args) {
		arg := cmd.args[i]
		i++

		parsersp := func() bool {
			if len(arg) < 2 || arg[0] != '@' {
				return false
			}
			response := cmd.parseResponse(arg[1:])
			cmd.args = append(cmd.args[:i], append(response, cmd.args[i:]...)...)
			return true
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
		    !parsearg("--dest=", &cmd.Dest) &&
		    !parsearg("--repo=", &cmd.Repo) &&
			!parsearg("--op=", &cmd.Op) &&
		    !parsearg("--vcs=", &cmd.Vcs) &&
			!parsebool("--print", &cmd.print) &&
			!parsebool("-v", &cmd.Verbose) &&
			!parsebool("--verbose", &cmd.Verbose) &&
			!parsebool("-h", &cmd.Help) &&
			!parsebool("--help", &cmd.Help) {
			usage(1)
		}

		if cmd.Op != "" {
			break
		}

		if cmd.Help {
			usage(0)
		}

		if parseDiagnostic && cmd.print && cmd.Verbose {
			fmt.Printf("(print) i=%d cmd.args=%s\n", i, cmd.args)
			cmd.print = false
		}
	}

	if parseDiagnostic && cmd.Verbose {
		fmt.Printf("exit cmd.args=%s\n", cmd.args[i:])
	}
	return cmd.args[i:]
}

// Given @<response-file>, open file <response-file> and add
// its lines to args; we assume each line is a single arg
func (p *Command) parseResponse(responsePath string) []string {
	args, err := gsos.FileReadLines(responsePath)
	if err != nil {
		usage(1)
	}

	return args
}
