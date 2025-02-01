package target

const (
	Phony              = ".PHONY"
	Suffixes           = ".SUFFIXES"
	Default            = ".DEFAULT"
	Precious           = ".PRECIOUS"
	Intermediate       = ".INTERMEDIATE"
	Notintermediate    = ".NOTINTERMEDIATE"
	Secondary          = ".SECONDARY"
	Secondexpansion    = ".SECONDEXPANSION"
	DeleteOnError      = ".DELETE_ON_ERROR"
	Ignore             = ".IGNORE"
	LowResolutionTime  = ".LOW_RESOLUTION_TIME"
	Silent             = ".SILENT"
	ExportAllVariables = ".EXPORT_ALL_VARIABLES"
	Notparallel        = ".NOTPARALLEL"
	Oneshell           = ".ONESHELL"
	Posix              = ".POSIX"
)

var Builtin = []string{
	Phony,
	Suffixes,
	Default,
	Precious,
	Intermediate,
	Notintermediate,
	Secondary,
	Secondexpansion,
	DeleteOnError,
	Ignore,
	LowResolutionTime,
	Silent,
	ExportAllVariables,
	Notparallel,
	Oneshell,
	Posix,
}
