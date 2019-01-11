// vcs-torture/gsos/helpers.go

package gsos

import (
	"bufio"
	"fmt"
	"os"
)

func DataToLines(data []byte) []string {

	// Turn into lines split on [CR]LF
	var output []string
	for len(data) > 0 {
		adv, token, err := bufio.ScanLines(data, true)
		if err != nil {
			break
		}
		output = append(output, string(token))
		data = data[adv:]
	}

	return output
}

func FileWriteLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, L := range lines {
		_, err = w.WriteString(fmt.Sprintf("%s\n", L))
		if err != nil {
			return err
		}
	}

	return nil
}

// FileReadLines returns data from a file as individual lines of text
func FileReadLines(path string) (lines []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	fs := bufio.NewScanner(f)
	for fs.Scan() {
		lines = append(lines, fs.Text())
	}
	err = fs.Err()
	return
}
