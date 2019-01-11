// vcs-torture/vcs/worktree.go

package vcs

import (
	"fmt"
	"strings"
)

// Worktree manages the files that can be added to a repository
// This generates a consistent but unique pathname
type Worktree struct {
	root string
	numFiles int
	filesPerDir int
	dirsPerDir int

	nthUniqueName int
	files []string
	dirs map[string]int
	dirplace []int
}

func (w *Worktree) Generate(callback func(i int, total int) bool) bool {
	w.files = make([]string, 0, w.numFiles)
	w.dirs = make(map[string]int)

	// Create our directory generator
	w.dirplace = make([]int, 1, 6)
	w.dirplace[0] = 0

	// Create the paths where we will put files
	for i := 0; i < w.numFiles; i++ {
		fname := w.uniqueName()
		dirpath := w.getDir()

		// Remember the dir (if there is one)
		path := fname
		if dirpath != "" {
			w.dirs[dirpath] = 1
			path = dirpath + "/" + fname
		}

		// Remember this file
		w.files = append(w.files, path)

		if callback != nil && callback(i, w.numFiles) {
			break
		}
	}

	return callback == nil || callback(w.numFiles, w.numFiles)
}

// Make a unique name (encoding a unique number)
var stringAtoms []string = []string{
	"at", "bi", "do", "ex", "fa", "go", "hi", "if", "ja", "ki", "lo", "me", "no", "of", "pi", "qi",
}

func (w *Worktree) uniqueName() string {
	nth := w.nthUniqueName
	w.nthUniqueName += 1

	// Make a unique name
	var fragments []string
	for nth >= 16 {
		fragments = append(fragments, stringAtoms[nth % 16])
		nth >>= 4
	}
	fragments = append(fragments, stringAtoms[nth])
	return strings.Join(fragments, "_")
}

// Make a directory path to hold a file, increasing depth as
// number of files increases. We put filesPerDir files per directory,
// and dirsPerDir sub-directories per directory. A typical number
// is 48 files and 16 dirs per tree
func (w *Worktree) getDir() string {

	// Create current directory path
	var dirparts []string
	for i := len(w.dirplace) - 1; i > 0 ; i-- {
		dirparts = append(dirparts, fmt.Sprintf("%c", w.dirplace[i]+'a'))
	}
	dirpath := strings.Join(dirparts, "/")

	// increment for next time
	incr := w.filesPerDir // for files, will drop to w.dirsPerDir after
	L := len(w.dirplace)
	for i := 0; i < L; i++ {
		w.dirplace[i] += 1
		if w.dirplace[i] < incr {
			break
		}
		incr = w.dirsPerDir

		w.dirplace[i] = 0
		if L == i+1 {
			w.dirplace = append(w.dirplace, 0)
			break
		}
	}

	return dirpath
}
