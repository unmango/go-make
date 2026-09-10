WORKING_DIR := $(shell pwd)
GO_SOURCES := $(shell find . -name '*.go' -print | sed 's|^\./||')
