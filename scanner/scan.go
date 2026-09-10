package scanner

import (
	"bytes"
)

// ScanTokens is a [bufio.SplitFunc] for a [bufio.Scanner] that
// scans for tokens supported by the make syntax.
func ScanTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	switch data[0] {
	case ' ':
		return 1, data[:1], nil
	case '\r':
		if len(data) == 1 && !atEOF {
			return 0, nil, nil // We need more info to make a decision
		}
		if len(data) > 1 && data[1] == '\n' {
			return 2, data[:2], nil // A CRLF line ending is a single token
		}

		return 1, data[:1], nil
	case '?', '+':
		// '?=' and '+=' are operators of their own. Neither character means
		// anything to make on its own, so one written without the '=' is
		// ordinary text and falls through to the search below.
		if len(data) > 1 && data[1] == '=' {
			return 2, data[:2], nil
		}
	case '!':
		if len(data) == 1 && !atEOF {
			return 0, nil, nil // We need more info to make a decision
		}
		if len(data) > 1 && data[1] == '=' {
			return 2, data[:2], nil
		}

		return 1, data[:1], nil
	case ':':
		if len(data) < 4 && !atEOF {
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
		fallthrough
	case '\n', '\t', '$', '(', ')', '{', '}', ',', '\'', '"', ';', '|', '=':
		return 1, data[:1], nil
	}

	if i := delimiterIndex(data); i > 0 {
		return i, data[:i], nil
	}

	if atEOF {
		return len(data), data, nil
	} else {
		return 0, nil, nil
	}
}

// delimiterIndex reports where the first delimiter in data is written, or -1
// when data holds none. It is the index the token starting at data[0] ends at.
func delimiterIndex(data []byte) int {
	i := bytes.IndexAny(data, ":\r\n\t (){},'\"$;|=!")
	if p := poundIndex(data); p >= 0 && (i < 0 || p < i) {
		return p
	}
	// A '+' or a '?' written against an '=' opens an operator of its own, so
	// the token ends in front of it rather than between the two characters:
	// "VAR+=x" is "VAR", "+=", "x".
	if i > 1 && data[i] == '=' && (data[i-1] == '+' || data[i-1] == '?') {
		i--
	}

	return i
}

// poundIndex reports where the first '#' that begins a comment in data is
// written, or -1 when data holds none.
//
// A backslash escapes the pound, so the character stands for itself there,
// and the backslash may itself be escaped. A '#' begins a comment exactly
// when the run of backslashes written in front of it has even length, so
// `a\#b` is the text make expands to "a#b" and `a\\#b` is the text `a\\`
// followed by a comment.
func poundIndex(data []byte) int {
	for i, b := range data {
		if b != '#' {
			continue
		}

		n := 0
		for j := i - 1; j >= 0 && data[j] == '\\'; j-- {
			n++
		}
		if n%2 == 0 {
			return i
		}
	}

	return -1
}
