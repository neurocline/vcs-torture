// vcs-torture/vcs/svn.go

package vcs

// Run a Subversion client command, returning elapsed time and stdout and stderr
func RunSvnCommand(repodir string, env []string, cmd ...string) (float64, []byte, []byte) {

	return RunExternal("svn", repodir, env, cmd...)
}

// Run a Subversion admin command, returning elapsed time and stdout and stderr
func RunSvnadminCommand(repodir string, env []string, cmd ...string) (float64, []byte, []byte) {

	return RunExternal("svnadmin", repodir, env, cmd...)
}
