package variable

const (
	Makefiles    = "MAKEFILES"
	Vpath        = "VPATH"
	Shell        = "SHELL"
	Makeshell    = "MAKESHELL"
	Make         = "MAKE"
	MakeVersion  = "MAKE_VERSION"
	MakeHost     = "MAKE_HOST"
	Makelevel    = "MAKELEVEL"
	Makeflags    = "MAKEFLAGS"
	Gnumakeflags = "GNUMAKEFLAGS"
	Makecmdgoals = "MAKECMDGOALS"
	Curdir       = "CURDIR"
	Suffixes     = "SUFFIXES"
	Libpatterns  = ".LIBPATTERNS"
)

var Special = []string{
	Makefiles,
	Vpath,
	Shell,
	Makeshell,
	Make,
	MakeVersion,
	MakeHost,
	Makelevel,
	Makeflags,
	Gnumakeflags,
	Makecmdgoals,
	Curdir,
	Suffixes,
	Libpatterns,
}
