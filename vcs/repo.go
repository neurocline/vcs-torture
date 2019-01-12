// vcs-torture/vcs/repo.go

package vcs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vcs-torture/gsos"
)

type OpParams struct {

	// Worktree parameters
	NumFiles int
	FilesPerDir int
	DirsPerDir int
	FileSize int
}

type Repo struct {
	verbose bool
	startTime time.Time
	maxCmdline int

	dest string
	repoName string
	vcs string

	repo string
	server string // for client/server systems like Subversion

	numCommits int
	numHeadFiles int
}

func NewRepo(dest string, repoName string, vcs string, startTime time.Time) *Repo {
	r := &Repo{dest: dest, repoName: repoName, vcs: vcs}
	r.repo = filepath.Join(dest, repoName)
	if r.vcs == "svn" {
		r.server = fmt.Sprintf("file:///%s-svnrepo", r.repo)
	}
	r.startTime = startTime

	return r
}

func (r *Repo) GetRepo() string {
	return r.repo
}

// DeleteRepo removes the repo (and associated data, e.g the actual repo
// for Subversion operations).
// TBD this is incredibly dangerous. We should probably write a
// log file and check it so someone doesn't accidentally delete their
// hard disk - we should only delete things we created. Suggest creating
// a .vcs-torture file in the --dest dir and then refusing to do anything
// if it doesn't exist, and only delete things mentioned in this file.
func DeleteRepo(dest string, repoName string, vcs string) bool {
	err := os.RemoveAll(filepath.Join(dest, repoName))
	if err == nil && vcs == "svn" {
		err = os.RemoveAll(filepath.Join(dest, repoName+"-svnrepo"))
	}
	return err == nil
}

func (r *Repo) SetVerbose(verbose bool) {
	r.verbose = true
}

// Create makes a new repo or gets information about an existing repo.
func (r *Repo) Create() bool {

	// If it already exists, query information about it
	if _, err := os.Stat(r.repo); err == nil {
		return r.loadInfo()
	}

	return r.createRepo()
}

// loadInfo fetches useful information about an existing repo
func (r *Repo) loadInfo() bool {

	// Git
	if r.vcs == "git" {
		// Get the number of files in the tip of the tree
		_, stdout, _ := RunGitCommand(r.repo, nil, "ls-tree", "-r", "HEAD")
		r.numHeadFiles = len(gsos.DataToLines(stdout))

		// Get the number of commits in the repo
		_, stdout, _ = RunGitCommand(r.repo, nil, "log", "--oneline")
		r.numCommits = len(gsos.DataToLines(stdout))

		return true
	}

	// Mercurial
	if r.vcs == "hg" {
		// Not supported yet
		return false
	}

	// Subversion
	if r.vcs == "svn" {
		// Not supported yet
		return false
	}

	// Unknown source control system
	return false
}

// createRepo creates a new empty repo
func (r *Repo) createRepo() bool {
	// Create repo directory
	err := os.Mkdir(r.repo, os.ModePerm)
	if err != nil {
		return false
	}

	// Git
	if r.vcs == "git" {
		// "git init"
		delta, stdout, stderr := RunGitCommand(r.repo, nil, "init")
		if r.verbose {
			fmt.Printf("T+%.2f: (elapsed=%.4f) git init\n", time.Since(r.startTime).Seconds(), delta)
			showStdoutStderr(stdout, stderr)
		}

		// "git config gc.auto 0" (turn off auto GC for this repo)
		delta, stdout, stderr = RunGitCommand(r.repo, nil, "config", "gc.auto", "0")
		if r.verbose {
			fmt.Printf("T+%.2f: (elapsed=%.4f) git config gc.auto 0\n", time.Since(r.startTime).Seconds(), delta)
			showStdoutStderr(stdout, stderr)
		}

		// "git config gc.autodetach false" for good measure
		delta, stdout, stderr = RunGitCommand(r.repo, nil, "config", "gc.autodetach", "false")
		if r.verbose {
			fmt.Printf("T+%.2f: (elapsed=%.4f) git config gc.autodetach false\n", time.Since(r.startTime).Seconds(), delta)
			showStdoutStderr(stdout, stderr)
		}
	}

	if r.vcs == "hg" {
		// "hg init"
		delta, stdout, stderr := RunHgCommand(r.repo, nil, "init")
		if r.verbose {
			fmt.Printf("T+%.2f: (elapsed=%.4f) hg init\n", time.Since(r.startTime).Seconds(), delta)
			showStdoutStderr(stdout, stderr)
		}
	}

	if r.vcs == "svn" {
		// "svnadmin create"
		delta, stdout, stderr := RunSvnadminCommand(r.dest, nil, "create", fmt.Sprintf("%s-svnrepo", r.repoName))
		if r.verbose {
			fmt.Printf("T+%.2f: (elapsed=%.4f) svnadmin create\n", time.Since(r.startTime).Seconds(), delta)
			showStdoutStderr(stdout, stderr)
		}
		// "svn checkout"
		delta, stdout, stderr = RunSvnCommand(r.dest, nil, "checkout", r.server, r.repo)
		if r.verbose {
			fmt.Printf("T+%.2f: (elapsed=%.4f) svnadmin create\n", time.Since(r.startTime).Seconds(), delta)
			showStdoutStderr(stdout, stderr)
		}
	}

	return true
}

func showStdoutStderr(stdout []byte, stderr []byte) {
	if len(stdout) != 0 {
		for _, v := range gsos.DataToLines(stdout) {
			fmt.Printf("%s\n", v)
		}
	}
	if len(stderr) != 0 {
		for _, v := range gsos.DataToLines(stderr) {
			fmt.Printf("(stderr): %s\n", v)
		}
	}
}
