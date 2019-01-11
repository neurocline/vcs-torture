// vcs-torture/vcs/git.go

package vcs

// Run a Git command, returning elapsed time and stdout and stderr
func RunGitCommand(repodir string, env []string, cmd ...string) (float64, []byte, []byte) {

	return RunExternal("git", repodir, env, cmd...)
}
