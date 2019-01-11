// vcs-torture/vcs/worktree.go

package vcs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Worktree manages the files that can be added to a repository
// This generates a consistent but unique pathname
type Worktree struct {
	verbose bool

	root string
	numFiles int
	filesPerDir int
	dirsPerDir int
	fileSize int

	nthUniqueName int
	nthContent int

	files []string
	dirs map[string]int
	dirplace []int
}

func NewWorktree(root string, numFiles int, filesPerDir int, dirsPerDir int, fileSize int) *Worktree {
	if numFiles == 0 {
		numFiles = 1000
	}
	if filesPerDir == 0 {
		filesPerDir = 48
	}
	if dirsPerDir == 0 {
		dirsPerDir = 16
	}
	if fileSize == 0 {
		fileSize = 10000
	}
	w := &Worktree{root:root, numFiles: numFiles, filesPerDir: filesPerDir, dirsPerDir: dirsPerDir, fileSize: fileSize}
	return w
}

func (w *Worktree) SetVerbose(verbose bool) {
	w.verbose = true
}

type WorktreeCallbackData struct {
	Done bool
	Pos int
	NumFiles int
	Path string
}

func (w *Worktree) Generate(callback func(cb *WorktreeCallbackData) bool) bool {
	w.files = make([]string, 0, w.numFiles)
	w.dirs = make(map[string]int)

	var cb WorktreeCallbackData
	cb.NumFiles = w.numFiles

	// Create our directory generator
	w.dirplace = make([]int, 1, 6)
	w.dirplace[0] = 0

	// Create the paths where we will put files
	for cb.Pos = 0; cb.Pos < w.numFiles; cb.Pos++ {
		fname := w.uniqueName()
		dirpath := w.getDir()

		// If we have a new dir, create it
		if _, ok := w.dirs[dirpath]; !ok {
			w.dirs[dirpath] = 1
			os.MkdirAll(filepath.Join(w.root, dirpath), os.ModePerm)
		}

		// Build the path (dir + name)
		cb.Path = fname
		if dirpath != "" {
			w.dirs[dirpath] = 1
			cb.Path = dirpath + "/" + fname
		}

		// Make this file if it doesn't already exist
		w.files = append(w.files, cb.Path)

		fpath := filepath.Join(w.root, cb.Path)
		if _, err := os.Stat(fpath); err == nil {
			w.nthContent += 1
		} else {
			content := w.makeContent(w.fileSize)
			err := ioutil.WriteFile(fpath, content, os.ModePerm)
			if err != nil {
				log.Fatalf("\nCouldn't write %s: %s\n", fpath, err)
			}
		}

		if callback != nil && callback(&cb) {
			break
		}
	}

	cb.Done = true
	return callback == nil || !callback(&cb)
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

// Make unique content of the specified size.
var contentAtoms []string = []string{
	"include", "for", "each", "int", "call", "lang", "operation", "overflow",
	"add", "multiply", "sub", "divide", "float", "array", "{", "}",
	"goto", "return", "range", "make", "byte", "var", "sizeof", "sink",
	"[", "]", "(", ")", "append", "copy", ":=", "==",
}
func (w *Worktree) makeContent(size int) []byte {
	nth := w.nthContent
	w.nthContent += 1

	crlf := false
	if runtime.GOOS == "windows" {
		crlf = true
	}

	content := make([]byte, size)
	col := 0
	for i := 0; i < size; {
		if col >= 100 && i < size {
			if crlf {
				content[i-2] = '\r'
			}
			content[i-1] = '\n'
			col = 0
		}
		var token []byte = []byte(contentAtoms[nth & 31] + " ")
		nth = ((nth << 27) | (nth >> 5)) + nth + 13
		L := len(token)
		if i+L > size {
			L = size - i
		}
		copy(content[i:i+L], token[:L])
		i += L
		col += L
	}

	if crlf {
		content[size-2] = '\r'
	}
	content[size-1] = '\n'

	if len(content) != size {
		log.Fatalf("Oops: %s\n", string(content))
	}
	return content
}
