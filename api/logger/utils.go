package logger

import (
	"bytes"
	"io"
)

func NewCapturingReadCloser(original io.ReadCloser) *CapturingReadCloser {
	return &CapturingReadCloser{
		original: original,
		captured: new(bytes.Buffer),
	}
}

func (crc *CapturingReadCloser) Read(p []byte) (n int, err error) {
	n, err = crc.original.Read(p)
	if n > 0 {
		crc.captured.Write(p[:n])
	}
	return
}

func (crc *CapturingReadCloser) Close() error {
	return crc.original.Close()
}
