// vcs-torture/gsos/helpers.go

package gsos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
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

type PeriodicStatus struct {
	startTime  time.Time
	lastStatus time.Time
	delta      time.Duration
	cols       int
}

func NewPeriodicStatus(startTime time.Time, delta time.Duration, cols int) *PeriodicStatus {
	if delta <= 0 {
		delta = 100 * time.Millisecond
	}
	if cols <= 0 {
		cols = 100
	}
	s := &PeriodicStatus{startTime: startTime, lastStatus: time.Now(), delta: delta, cols: cols}
	return s
}

func (s *PeriodicStatus) Ready() bool {
	return time.Since(s.lastStatus) >= s.delta
}

func (s *PeriodicStatus) Show(prompt string) {
	out := fmt.Sprintf("T+%.2f: %s", time.Since(s.startTime).Seconds(), prompt)

	padlen := s.cols - 1 - len(out)
	if padlen > 0 {
		out = out + strings.Repeat(" ", padlen)
	}

	fmt.Fprintf(os.Stderr, "\r%s", out)
	s.lastStatus = time.Now()
}
