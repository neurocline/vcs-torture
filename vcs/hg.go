// vcs-torture/vcs/hg.go

package vcs

// Run a Mercurial command, returning elapsed time and stdout and stderr
func RunHgCommand(repodir string, env []string, cmd ...string) (float64, []byte, []byte) {

	return RunExternal("hg", repodir, env, cmd...)
}
