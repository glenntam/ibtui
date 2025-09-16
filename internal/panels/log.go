package panels

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
)

// Given a file and a byte location (pos), return
// the location of the previous occurrence of "\n".
// Return itself if there are none.
func PrevNewline(f *os.File, pos int64) int64 {
	if pos <= 0 {
		return pos
	}
	fileSize := GetFileSize(f)
	if pos > fileSize {
		return fileSize
	}
	buf := make([]byte, 1)
	cur := pos - 1
	for cur > 0 {
		cur--
		_, err := f.ReadAt(buf, cur)
		if err != nil {
			return pos
		}
		if buf[0] == '\n' {
			return cur
		}
	}
	return pos
}

// Given a file and a byte location (pos), return
// the location of the next occurrence of "\n".
// Return itself if there are none.
func NextNewline(f *os.File, pos int64) int64 {
	if pos <= 0 {
		return pos
	}
	fileSize := GetFileSize(f)
	if pos >= fileSize {  // Last line already
		return fileSize
	}
	buf := make([]byte, 1)
	cur := pos
	for cur <= fileSize {
		cur++
		_, err := f.ReadAt(buf, cur)
		if err != nil {
			return pos
		}
		if buf[0] == '\n' {
			cur++
			return cur
		}
	}
	return pos
}

// Check if a given cursor location is at the end of file.
func CursorAtEOF(f *os.File, cur int64) bool {
	fileSize := GetFileSize(f)
	if cur >= 0 && cur < fileSize { // within valid bounds of the file
		return false
	}
	return true
}

// Retrieve a file's size. If unable return 0.
func GetFileSize(f *os.File) int64 {
	var fileSize int64
	fileInfo, err := f.Stat()
	if err != nil {
		slog.Error("Couldn't retrieve file size", "error", err)
	} else {
		fileSize = fileInfo.Size()
	}
	return fileSize
}

// Return the line given the position of the file.
func getLine(f *os.File, pos int64) string {
	prev := PrevNewline(f, pos)
	line := make([]byte, pos-prev)
	_, err := f.ReadAt(line, prev)
	if err != nil {
		return ""
	}
	return string(line)
}

// Return a slice of strings of a file based on
// the end position and number of lines desired.
func RenderLog(f *os.File, pos int64, n, w int) []string {
	var result []string
	for i := 0; i < n; i++ {
		l := getLine(f, pos)
		lineObj := make(map[string]any)
		err := json.Unmarshal([]byte(l), &lineObj) // Parse the individual log entry
		if err != nil {
			slog.Error("Couldn't Unmarshal JSON line object", "error", err, "logline", l)
			break
		}

		level, ok := lineObj["level"].(string)
		if !ok {
			slog.Error("Log entry in log file didn't have a log level", "error", err, "logline", l)
			break
		} else {
			delete (lineObj, "level")
		}

		timestamp, ok := lineObj["time"].(string)
		if !ok {
			slog.Error("Log entry in log file didn't have a timestamp", "error", err, "logline", l)
			break
		} else {
			delete (lineObj, "time")
		}
		t, err := time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, timestamp)
		}
		timeStr := t.Format("Jan 02, 15:04")

		msg, ok := lineObj["msg"].(string)
		if ok {
			delete (lineObj, "msg")
		}

		keys := make([]string, 0, len(lineObj))
		for k := range lineObj {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		objs := make([]string, 0, len(keys))
		for _, key := range keys {
			objs = append(objs, fmt.Sprintf("%s=%v", key, lineObj[key]))
		}
		logDisplayStr := fmt.Sprintf("%s [%s] %s. %s", level, timeStr, msg, strings.Join(objs, ", "))

		result = append(result, logDisplayStr)
		pos = PrevNewline(f, pos) // Started at the bottom; move up to the previous line
	}
	slices.Reverse(result)

	var wrapped []string
	for _, s := range result {
		for len(s) > w {
			wrapped = append(wrapped, s[:w])
			s = s[w:]
		}
		if len(s) > 0 {
			wrapped = append(wrapped, s)
		}
	}
	return wrapped[len(wrapped)-n:]
}
