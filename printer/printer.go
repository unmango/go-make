package printer

import (
	"go/token"
	"io"
)

type printer struct {
	w       io.Writer
	lastTok token.Token
	pos     token.Pos
	lastPos token.Pos
}
