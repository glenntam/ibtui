// Package typewriter tracks cursor movement in a file. Used for scrolling display of a log file.
package typewriter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ErrNegativePosition occurs when a file operation is given a negative position.
var ErrNegativePosition = errors.New("file position can't be negative")

// PrevNewline scans backwards from byte location "pos" and returns
// the byte location of the previous occurrence of "\n".
//
// Returns itself if there are no previous occurrences of "\n".
func PrevNewline(f *os.File, pos int64) (int64, error) {
	if pos <= 0 {
		return pos, ErrNegativePosition
	}
	fileSize, err := GetFileSize(f)
	if err != nil {
		return pos, fmt.Errorf("couldn't get previous newline because couldn't get file size: %w", err)
	}
	if pos > fileSize {
		return fileSize, nil
	}
	buf := make([]byte, 1)
	cur := pos - 1
	for cur > 0 {
		cur--
		_, err := f.ReadAt(buf, cur)
		if err != nil {
			return pos, fmt.Errorf("couldn't get previous newline because couldn't read file at position: %w", err)
		}
		if buf[0] == '\n' {
			return cur, nil
		}
	}
	return pos, nil
}

// NextNewline scans forwards from byte location "pos" and returns the byte location of the next occurrence of "\n".
// Return itself if there are no subsequent occurrences of "\n".
func NextNewline(f *os.File, pos int64) (int64, error) {
	if pos <= 0 {
		return pos, ErrNegativePosition
	}
	fileSize, err := GetFileSize(f)
	if err != nil {
		return pos, fmt.Errorf("couldn't get next newline because couldn't get file size: %w", err)
	}
	if pos >= fileSize { // Last line already
		return fileSize, nil
	}
	buf := make([]byte, 1)
	cur := pos
	for cur <= fileSize {
		cur++
		_, err := f.ReadAt(buf, cur)
		if err != nil {
			return pos, fmt.Errorf("couldn't get next newline because couldn't read file at position: %w", err)
		}
		if buf[0] == '\n' {
			cur++
			return cur, nil
		}
	}
	return pos, nil
}

// CursorAtEOF checks if a given cursor location is at the end of file.
func CursorAtEOF(f *os.File, cur int64) (bool, error) {
	fileSize, err := GetFileSize(f)
	if err != nil {
		return false, fmt.Errorf("couldn't determine cursor location because couldn't get file size: %w", err)
	}
	if cur >= 0 && cur < fileSize { // within valid bounds of the file
		return false, nil
	}
	return true, nil
}

// GetFileSize retrieves a file's size. If unable return 0.
func GetFileSize(f *os.File) (int64, error) {
	var fileSize int64
	fileInfo, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("couldn't get file size because couldn't stat log file: %w", err)
	}
	fileSize = fileInfo.Size()
	return fileSize, nil
}

// Return the line given the position of the file.
func getLine(f *os.File, pos int64) (string, error) {
	prev, err := PrevNewline(f, pos)
	if err != nil {
		return "", fmt.Errorf("couldn't get line number because couldn't find previous new line: %w", err)
	}
	line := make([]byte, pos-prev)
	_, err = f.ReadAt(line, prev)
	if err != nil {
		return "", fmt.Errorf("couldn't get line number because couldn't read line at cursor location: %w", err)
	}
	return string(line), nil
}

// RenderLog returns a slice of strings of a file based on a given end position and number of lines desired.
func RenderLog(f *os.File, pos int64, n, w int) ([]string, error) {
	result := make([]string, n)
	for i := n - 1; i >= 0; i-- {
		l, err := getLine(f, pos)
		if err != nil {
			result[i] = "-"
			continue
		}
		lineObj := make(map[string]any)
		_ = json.Unmarshal([]byte(l), &lineObj) // Parse the individual log entry
		// Separately handle "level"
		level, ok := lineObj["level"].(string)
		if !ok {
			result[i] = "-"
			continue
		}
		delete(lineObj, "level")
		// Separately handle "time"
		timestamp, ok := lineObj["time"].(string)
		if !ok {
			result[i] = "-"
			continue
		}
		delete(lineObj, "time")
		t, err := time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			t, err = time.Parse(time.RFC3339, timestamp)
			if err != nil {
				result[i] = "-"
				continue
			}
		}
		timeStr := t.Format("Jan 02, 15:04")
		// Separately handle "msg", if it exists
		msg, ok := lineObj["msg"].(string)
		if ok {
			delete(lineObj, "msg")
		}
		// Flatten all other keys into a string
		keys := make([]string, 0, len(lineObj))
		for k := range lineObj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		objs := make([]string, 0, len(keys))
		for _, key := range keys {
			objs = append(objs, fmt.Sprintf("%s=%v", key, lineObj[key]))
		}
		result[i] = fmt.Sprintf(
			"%s [%s] %s. %s",
			level,
			timeStr,
			msg,
			strings.Join(objs, ", "),
		)
		// Compile whole log entry string
		pos, err = PrevNewline(f, pos) // Started at the bottom; move up to the previous line
		if err != nil {
			break
		}
	}
	// Truncate the strings to the screen width
	for i := range result {
		r := []rune(result[i])
		if len(r) > w {
			result[i] = string(r[:w])
		}
	}
	return result, nil
}
