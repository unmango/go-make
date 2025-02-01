package make

import (
	"github.com/unmango/go-make/parser"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/scanner"
	"github.com/unmango/go-make/writer"
)

var (
	Fprint     = printer.Fprint
	NewParser  = parser.New
	NewScanner = scanner.New
	NewWriter  = writer.New
	ScanTokens = scanner.ScanTokens
)
