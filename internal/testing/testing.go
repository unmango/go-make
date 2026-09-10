package testing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

// ChunkReader reads Data in chunks of at most Size bytes per call to Read.
type ChunkReader struct {
	Data []byte
	Size int
	pos  int
}

// NewChunkReader returns a ChunkReader that yields at most size bytes of data per read.
func NewChunkReader(data string, size int) *ChunkReader {
	return &ChunkReader{Data: []byte(data), Size: size}
}

func (c *ChunkReader) Read(p []byte) (int, error) {
	if c.pos >= len(c.Data) {
		return 0, io.EOF
	}

	n := min(c.Size, min(len(p), len(c.Data)-c.pos))
	copy(p, c.Data[c.pos:c.pos+n])
	c.pos += n

	return n, nil
}
