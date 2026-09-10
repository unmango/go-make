# A makefile that uses CRLF line endings.
VAR := val

all: build test

build: $(VAR)
	echo building $(VAR)

test:
	echo testing
