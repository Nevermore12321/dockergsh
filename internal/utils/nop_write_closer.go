package utils

import "io"

type NopWriter struct{}

func (NopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error {
	return nil
}

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}
