package internal

import (
	"bytes"
	"runtime"
	"strconv"
)

const stackPageSize = 16

func Stack(skip int) (stack []byte, file string, line int) {
	skip += 2 // 0:runtime.Callers, 1:this(Stack())
	sb := bytes.Buffer{}
	for {
		list := make([]uintptr, stackPageSize)
		n := runtime.Callers(skip, list)
		if n > 0 {
			frames := runtime.CallersFrames(list)
			for frame, more := frames.Next(); more; frame, more = frames.Next() {
				if sb.Len() != 0 {
					sb.WriteByte('\n')
				} else {
					file = frame.File
					line = frame.Line
				}
				sb.WriteString(frame.Function)
				sb.WriteByte('\n')
				sb.WriteByte('\t')
				sb.WriteString(frame.File)
				sb.WriteByte(':')
				sb.WriteString(strconv.Itoa(frame.Line))
			}
			skip += n
		}
		if n < stackPageSize {
			break
		}
	}
	return sb.Bytes(), file, line
}
