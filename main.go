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
	"strconv"
	"time"

	"vcs-torture/gsos"
	"vcs-torture/vcs"
)

// main parses all the command-line arguments, running in
// groups, one group for each distinct command
func main() {
	cmd := &Command{args: os.Args[1:]}
	cmd.startTime = time.Now()

	for len(cmd.args) > 0 {
		cmd.Op = ""
		cmd.args = cmd.parse()
		cmd.Run()
	}
}

// ----------------------------------------------------------------------------------------------

func (cmd *Command) Run() {
	if parseDiagnostic && cmd.Verbose {
		fmt.Printf("cmd: %+v\n", cmd)
	}

	if cmd.Verbose {
		fmt.Printf("op=%s\n", cmd.Op)
	}

	switch cmd.Op {
	case "create":
		cmd.OpCreate()
	case "remove":
		cmd.OpRemove()
	case "worktree":
		cmd.OpWorktree()
	default:
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

func (cmd *Command) OpCreate() {
	cmd.mustHaveDest()
	cmd.mustHaveRepo()
	cmd.mustHaveVcs()

	repo := vcs.NewRepo(cmd.Dest, cmd.Repo, cmd.Vcs, cmd.startTime)
	repo.SetVerbose(cmd.Verbose)
	if !repo.Create() {
		log.Fatalf("Couldn't create repo\n")
	}
}

func (cmd *Command) OpRemove() {
	if !vcs.DeleteRepo(cmd.Dest, cmd.Repo, cmd.Vcs) {
		log.Fatalf("Couldn't remove repo\n")
	}
}

func (cmd *Command) OpWorktree() {
	cmd.mustHaveDest()
	cmd.mustHaveRepo()

	w := vcs.NewWorktree(cmd.Dest, cmd.Repo, cmd.Params)
	w.SetVerbose(cmd.Verbose)

	status := gsos.NewPeriodicStatus(cmd.startTime, 20*time.Millisecond, 0)
	fn := func(cb *vcs.WorktreeCallbackData) bool {
		if cmd.Abort {
			return true
		}
		if !cb.Done && !status.Ready() {
			return false
		}
		fnamedisp := cb.Path
		if len(fnamedisp) > 39 {
			fnamedisp = cb.Path[:18]+"..." + cb.Path[len(cb.Path)-18:]
		}
		status.Show(fmt.Sprintf("create %d/%d files: %s", cb.Pos, cb.NumFiles, fnamedisp))
		if cb.Done {
			fmt.Fprintf(os.Stderr, "\n")
		}
		return false
	}

	if !w.Generate(fn) {
		log.Fatalf("Couldn't put files in worktree\n")
	}
}

// ----------------------------------------------------------------------------------------------

type Command struct {
	// Op is what we are doing (create, add, ...)
	Op string

	// Vcs is the name of the version control system to test
	Vcs string

	// Dest points to our working area
	Dest string

	// Repo is the path to the repo to work on (inside Dest)
	Repo string

	// Params is the set of Op values that can be set on the command-line
	Params vcs.OpParams

	Help    bool
	Verbose bool
	Abort bool

	// remaining command-line arguments
	args []string

	print bool
	startTime time.Time
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

		parseint := func(opt string, val *int) bool { return ParseIntArg(arg, opt, val) }
		parsestr := func(opt string, val *string) bool { return ParseStrArg(arg, opt, val) }
		parsebool := func(opt string, val *bool) bool { return ParseBoolArg(arg, opt, val) }

		if !parsersp() &&
		    !parsestr("--dest=", &cmd.Dest) &&
		    !parsestr("--repo=", &cmd.Repo) &&
			!parsestr("--op=", &cmd.Op) &&
		    !parsestr("--vcs=", &cmd.Vcs) &&

		    !parseint("--worktree-file-count=", &cmd.Params.NumFiles) &&
		    !parseint("--worktree-file-size=", &cmd.Params.FileSize) &&
		    !parseint("--files-per-dir=", &cmd.Params.FilesPerDir) &&
		    !parseint("--dirs-per-dir=", &cmd.Params.DirsPerDir) &&

			!parsebool("--print", &cmd.print) &&
			!parsebool("-v", &cmd.Verbose) &&
			!parsebool("--verbose", &cmd.Verbose) &&
			!parsebool("-h", &cmd.Help) &&
			!parsebool("--help", &cmd.Help) {
			fmt.Printf("Don't understand '%s'\n", arg)
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

func ParseIntArg(arg string, opt string, val *int) bool {
	var strval string
	if !ParseStrArg(arg, opt, &strval) {
		return false
	}

	n, err := strconv.Atoi(strval)
	if err != nil {
		return false
	}
	*val = n
	return true
}

func ParseStrArg(arg string, opt string, val *string) bool {
	optlen := len(opt)
	if len(arg) <= optlen || arg[:optlen] != opt {
		return false
	}
	*val = arg[optlen:]
	return true
}

func ParseBoolArg(arg string, opt string, val *bool) bool {
	if arg != opt {
		return false
	}
	*val = true
	return true
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
