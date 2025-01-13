package make

import (
	"bytes"
)

// ScanTokens is a [bufio.SplitFunc] for a [bufio.Scanner] that
// scans for tokens supported by the make syntax.
func ScanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	switch data[0] {
	case ' ':
		return 1, nil, nil // TODO: Treat this as a token?
	case '?':
		if len(data) > 1 && data[1] == '=' {
			return 2, data[:2], nil
		}
	case ':':
		if len(data) == 1 && !atEOF {
			return 0, nil, nil // We need more info to make a decision
		}
		if bytes.HasPrefix(data, []byte(":::=")) {
			return 4, data[:4], nil
		}
		if bytes.HasPrefix(data, []byte("::=")) {
			return 3, data[:3], nil
		}
		if bytes.HasPrefix(data, []byte(":=")) {
			return 2, data[:2], nil
		}

		fallthrough
	case '#':
		if len(data) > 1 && data[1] == ' ' {
			return 2, data[:1], nil
		}

		fallthrough
	case '\n', '\t', '$', '(', ')', '{', '}', ',':
		return 1, data[:1], nil
	}

	if i := bytes.IndexAny(data, ":\n\t (){},"); i > 0 {
		switch data[i] {
		case ' ':
			return i + 1, data[:i], nil
		case ':', '\n', '\t', '(', ')', '{', '}', ',':
			return i, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	} else {
		return 0, nil, nil
	}
}
