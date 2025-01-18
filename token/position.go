package token

import "go/token"

type (
	Pos      = token.Pos
	Position = token.Position
	File     = token.File
	FileSet  = token.FileSet
)

const NoPos = token.NoPos

// PositionFor returns the Position value for the given file position p.
// If p is out of bounds, it is adjusted to match the File.Offset behavior.
// p must be a Pos value in file or NoPos. Calling token.PositionFor(file, p)
// is equivalent to calling file.PositionFor(p, false).
func PositionFor(file *File, p Pos) Position {
	return file.PositionFor(p, false)
}
