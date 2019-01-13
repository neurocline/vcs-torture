// vcs-torture/vcs/repo.go

package vcs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"vcs-torture/gsos"
)

type RepoOptions struct {
	NumCommits    int
	AddsPerCommit int
	FilesPerAdd   int
}

type Repo struct {
	RepoOptions

	verbose    bool
	startTime  time.Time
	maxCmdline int

	dest     string
	repoName string
	vcs      string
	overhead float64

	repo   string
	server string // for client/server systems like Subversion

	numCommits   int
	numHeadFiles int

	Worktree *Worktree
}

func NewRepo(dest string, repoName string, vcs string, startTime time.Time, options RepoOptions) *Repo {
	r := &Repo{dest: dest, repoName: repoName, vcs: vcs, RepoOptions: options}
	r.repo = filepath.Join(dest, repoName)
	if r.vcs == "svn" {
		r.server = fmt.Sprintf("file:///%s-svnrepo", r.repo)
	}
	r.startTime = startTime

	// Set up command line limit (these are puposely much lower than
	// the real limits)
	if runtime.GOOS == "windows" {
		r.maxCmdline = 2000
	} else {
		r.maxCmdline = 8000
	}

	return r
}

func (r *Repo) GetRepo() string {
	return r.repo
}

func (r *Repo) AddWorktree(options WorktreeOptions) {
	r.Worktree = NewWorktree(r.dest, r.repoName, options)
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

// ----------------------------------------------------------------------------------------------

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
			fmt.Printf("T+%.2f: (elapsed=%.4f) svn checkout\n", time.Since(r.startTime).Seconds(), delta)
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

// ----------------------------------------------------------------------------------------------

type CommitCallbackData struct {
	Done          bool
	Commit        int
	NumIndexFiles int
	LooseObjects  int
	PackObjects   int
}

func (r *Repo) Commit(callback func(cb *CommitCallbackData) bool) bool {
	//fmt.Printf("(*Repo).Commit\n")
	var cb CommitCallbackData

	var sumAddTime float64
	var sumCommitTime float64
	pos := 0
	for cb.Commit = 1; cb.Commit <= r.NumCommits; cb.Commit++ {

		// Add files for our commit
		numToAdd := r.AddsPerCommit * r.FilesPerAdd
		add := 0
		for add < numToAdd {
			amt := r.FilesPerAdd
			if add+amt > numToAdd {
				amt = numToAdd - add
			}
			addList := r.getFileSubset(pos+add, amt)
			amt = len(addList)
			deltaAdd := r.addFiles(addList) - r.overhead
			sumAddTime += deltaAdd

			add += amt
			cb.NumIndexFiles += amt
			if callback != nil && callback(&cb) {
				break
			}
		}

		pos += add

		// Now make the commit
		deltaCommit := r.makeCommit(cb.Commit) - r.overhead
		sumCommitTime += deltaCommit

		if callback != nil && callback(&cb) {
			break
		}
	}

	cb.Done = true
	return callback == nil || !callback(&cb)
}

// Get some files
func (r *Repo) getFileSubset(pos, amt int) []string {
	//fmt.Printf("(*Repo).getFileSubset\n")
	if pos+amt > len(r.Worktree.Files) {
		amt = len(r.Worktree.Files) - pos
	}
	addList := make([]string, amt)
	copy(addList, r.Worktree.Files[pos:pos+amt])
	return addList
}

// Do "git add" on this set of files. We may need to break this
// up into more than one command-line invocation
func (r *Repo) addFiles(addList []string) float64 {
	//fmt.Printf("(*Repo).addFiles\n")

	var addElapsed float64
	for start := 0; start < len(addList); {

		filelist := make([]string, 0, 100)
		cmdsize := 0

		for i := start; i < len(addList); i++ {
			path := addList[i]
			if cmdsize+1+len(path) > r.maxCmdline {
				break
			}
			filelist = append(filelist, path)
			cmdsize += 1 + len(path)
		}

		// Add this subset
		var delta float64
		if r.vcs == "git" {
			delta, _, _ = RunGitCommand(r.repo, nil, append([]string{"add"}, filelist...)...)
		} else if r.vcs == "hg" {
			delta, _, _ = RunHgCommand(r.repo, nil, append([]string{"add"}, filelist...)...)
		} else if r.vcs == "svn" {
			delta, _, _ = RunSvnCommand(r.repo, nil, append([]string{"add", "--parents" }, filelist...)...)
		}
		addElapsed += delta

		start += len(filelist)
	}

	return addElapsed
}

// Do "git commit" on the current repo (which should have files added to it)
func (r *Repo) makeCommit(n int) float64 {
	//fmt.Printf("(*Repo).makeCommit\n")
	var elapsed float64
	if r.vcs == "git" {
		elapsed, _, _ = RunGitCommand(r.repo, nil, "commit", "-m", fmt.Sprintf("commit %d", n))
	} else if r.vcs == "hg" {
		elapsed, _, _ = RunHgCommand(r.repo, nil, "commit", "-m", fmt.Sprintf("commit %d", n))
	} else if r.vcs == "svn" {
		elapsed, _, _ = RunSvnCommand(r.repo, nil, "commit", "-m", fmt.Sprintf("commit %d", n))
	}
	return elapsed
}
