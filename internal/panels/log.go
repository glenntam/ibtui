package panels

import (
	"os"
	"slices"
)


// Given a file and a byte location (pos), return
// the location of the previous occurrence of "\n".
// Return itself if there are none.
func PrevNewline(f *os.File, pos int64) int64 {
	if pos <= 0 {
		return pos
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return pos
	}
	fileSize := fileInfo.Size()
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
	fileInfo, err := f.Stat()
	if err != nil {
		return pos
	}
	fileSize := fileInfo.Size()
	if pos >= fileSize {  // Last line already
		return pos
	}
	buf := make([]byte, 1)
	cur := pos
	for cur < fileSize {
		cur++
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

// Return the line given the position of the file
func GetLine(f *os.File, pos int64) string {
	prev := PrevNewline(f, pos)
	line := make([]byte, pos-prev)
	_, err := f.ReadAt(line, prev)
	if err != nil {
		return ""
	}
	return string(line)
}

func RenderLog(f *os.File, pos int64, n int) []string {
	var res []string
	for i := 0; i < n; i++ {
		res = append(res, GetLine(f, pos))
		pos = PrevNewline(f, pos)
	}
	slices.Reverse(res)
	return res
}
