package panels

import (
	"log/slog"
	"os"
	"slices"
	"strings"
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
func GetLine(f *os.File, pos int64) string {
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
func RenderLog(f *os.File, pos int64, n int) []string {
	var res []string
	for i := 0; i < n; i++ {
		res = append(res, GetLine(f, pos))
		pos = PrevNewline(f, pos)
	}
	slices.Reverse(res)
	res[0] = strings.TrimLeft(res[0], "\r\n")
	res[len(res)-1] = strings.TrimRight(res[len(res)-1], "\r\n")
	return res
}
