package make

import (
	"github.com/unmango/go-make/parser"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/scanner"
	"github.com/unmango/go-make/writer"
)

const (
	MakefilesVariable    = "MAKEFILES"
	VpathVariable        = "VPATH"
	ShellVariable        = "SHELL"
	MakeshellVariable    = "MAKESHELL"
	MakeVariable         = "MAKE"
	MakeVersionVariable  = "MAKE_VERSION"
	MakeHostVariable     = "MAKE_HOST"
	MakelevelVariable    = "MAKELEVEL"
	MakeflagsVariable    = "MAKEFLAGS"
	GnumakeflagsVariable = "GNUMAKEFLAGS"
	MakecmdgoalsVariable = "MAKECMDGOALS"
	CurdirVariable       = "CURDIR"
	SuffixesVariable     = "SUFFIXES"
	LibpatternsVariable  = ".LIBPATTERNS"
)

var SpecialVariables = []string{
	MakefilesVariable,
	VpathVariable,
	ShellVariable,
	MakeshellVariable,
	MakeVariable,
	MakeVersionVariable,
	MakeHostVariable,
	MakelevelVariable,
	MakeflagsVariable,
	GnumakeflagsVariable,
	MakecmdgoalsVariable,
	CurdirVariable,
	SuffixesVariable,
	LibpatternsVariable,
}

const (
	PhonyTarget              = ".PHONY"
	SuffixesTarget           = ".SUFFIXES"
	DefaultTarget            = ".DEFAULT"
	PreciousTarget           = ".PRECIOUS"
	IntermediateTarget       = ".INTERMEDIATE"
	NotintermediateTarget    = ".NOTINTERMEDIATE"
	SecondaryTarget          = ".SECONDARY"
	SecondexpansionTarget    = ".SECONDEXPANSION"
	DeleteOnErrorTarget      = ".DELETE_ON_ERROR"
	IgnoreTarget             = ".IGNORE"
	LowResolutionTimeTarget  = ".LOW_RESOLUTION_TIME"
	SilentTarget             = ".SILENT"
	ExportAllVariablesTarget = ".EXPORT_ALL_VARIABLES"
	NotparallelTarget        = ".NOTPARALLEL"
	OneshellTarget           = ".ONESHELL"
	PosixTarget              = ".POSIX"
)

var BuiltinTargets = []string{
	PhonyTarget,
	SuffixesTarget,
	DefaultTarget,
	PreciousTarget,
	IntermediateTarget,
	NotintermediateTarget,
	SecondaryTarget,
	SecondexpansionTarget,
	DeleteOnErrorTarget,
	IgnoreTarget,
	LowResolutionTimeTarget,
	SilentTarget,
	ExportAllVariablesTarget,
	NotparallelTarget,
	OneshellTarget,
	PosixTarget,
}

var (
	Fprint     = printer.Fprint
	NewParser  = parser.New
	NewScanner = scanner.New
	NewWriter  = writer.New
	ScanTokens = scanner.ScanTokens
)
