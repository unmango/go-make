package testing

import (
	"bytes"
	"errors"
	"fmt"
)

type ErrReader string

func (e ErrReader) Read([]byte) (int, error) {
	return 0, errors.New(string(e))
}

type ErrWriter string

func (e ErrWriter) Write(p []byte) (int, error) {
	return 0, errors.New(string(e))
}

type ErrAfterWriter struct {
	After int
	Buf   *bytes.Buffer
	at    int
}

func NewErrAfterWriter(after int) *ErrAfterWriter {
	return &ErrAfterWriter{
		After: after,
		Buf:   &bytes.Buffer{},
	}
}

func (e *ErrAfterWriter) Write(p []byte) (int, error) {
	if e.at++; e.at >= e.After {
		return 0, fmt.Errorf("write err: %d", e.at)
	} else {
		return e.Buf.Write(p)
	}
}
