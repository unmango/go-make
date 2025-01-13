package testing

import "errors"

type ErrReader string

func (e ErrReader) Read([]byte) (int, error) {
	return 0, errors.New(string(e))
}

type ErrWriter string

func (e ErrWriter) Write(p []byte) (int, error) {
	return 0, errors.New(string(e))
}
