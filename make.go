package make

const (
	DefineDirective      = "define"
	EndefDirective       = "endef"
	UndefineDirective    = "undefine"
	IfdefDirective       = "ifdef"
	IfndefDirective      = "ifndef"
	IfeqDirective        = "ifeq"
	IfneqDirective       = "ifneq"
	ElseDirective        = "else"
	EndifDirective       = "endif"
	IncludeDirective     = "include"
	DashIncludeDirective = "-include"
	SincludeDirective    = "sinclude"
	OverrideDirective    = "override"
	ExportDirective      = "export"
	UnexportDirective    = "unexport"
	PrivateDirective     = "private"
	VpathDirective       = "vpath"
)

var Directives = []string{
	DefineDirective,
	EndefDirective,
	UndefineDirective,
	IfdefDirective,
	IfndefDirective,
	IfeqDirective,
	IfneqDirective,
	ElseDirective,
	EndifDirective,
	IncludeDirective,
	DashIncludeDirective,
	SincludeDirective,
	OverrideDirective,
	ExportDirective,
	UndefineDirective,
	PrivateDirective,
	VpathDirective,
}

const (
	SubstFunction      = "subst"
	PatsubstFunction   = "patsubst"
	StripFunction      = "strip"
	FindstringFunction = "findstring"
	FilterFunction     = "filter"
	FilterOutFunction  = "filter-out"
	SortFunction       = "sort"
	WordFunction       = "word"
	WordsFunction      = "words"
	WordlistFunction   = "wordlist"
	FirstwordFunction  = "firstword"
	LastwordFunction   = "lastword"
	DirFunction        = "dir"
	NotdirFunction     = "notdir"
	SuffixFunction     = "suffix"
	BasenameFunction   = "basename"
	AddsuffixFunction  = "addsuffix"
	AddprefixFunction  = "addprefix"
	JoinFunction       = "join"
	WildcardFunction   = "wildcard"
	RealpathFunction   = "realpath"
	AbspathFunction    = "abspath"
	ErrorFunction      = "error"
	WarningFunction    = "warning"
	ShellFunction      = "shell"
	OriginFunction     = "origin"
	FlavorFunction     = "flavor"
	LetFunction        = "let"
	ForeachFunction    = "foreach"
	IfFunction         = "if"
	OrFunction         = "or"
	AndFunction        = "and"
	IntcmpFunction     = "intcmp"
	CallFunction       = "call"
	EvalFunction       = "eval"
	FileFunction       = "file"
	ValueFunction      = "value"
)

var BuiltinFunctions = []string{
	SubstFunction,
	PatsubstFunction,
	StripFunction,
	FindstringFunction,
	FilterFunction,
	FilterOutFunction,
	SortFunction,
	WordFunction,
	WordsFunction,
	WordlistFunction,
	FirstwordFunction,
	LastwordFunction,
	DirFunction,
	NotdirFunction,
	SuffixFunction,
	BasenameFunction,
	AddsuffixFunction,
	AddprefixFunction,
	JoinFunction,
	WildcardFunction,
	RealpathFunction,
	AbspathFunction,
	ErrorFunction,
	WarningFunction,
	ShellFunction,
	OriginFunction,
	FlavorFunction,
	LetFunction,
	ForeachFunction,
	IfFunction,
	OrFunction,
	AndFunction,
	IntcmpFunction,
	CallFunction,
	EvalFunction,
	FileFunction,
	ValueFunction,
}

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

type (
	Target string
	PreReq string
	Recipe string
)

type Rule struct {
	Target  []string
	PreReqs []string
	Recipe  []string
}

type Makefile struct {
	Rules []Rule
}
