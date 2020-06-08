package sse

import "io"

type stringWriter interface {
    // 接口内嵌
	io.Writer
	WriteString(string) (int, error)
}

type stringWrapper struct {
	io.Writer
}

func (w stringWrapper) WriteString(str string) (int, error) {
	return w.Writer.Write([]byte(str))
}

func checkWriter(writer io.Writer) stringWriter {
	// writer是stringWriter类型
	if w, ok := writer.(stringWriter); ok {
		return w
	} else {
		return stringWrapper{writer}
	}
}
