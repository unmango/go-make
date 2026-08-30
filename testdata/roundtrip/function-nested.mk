SOURCES := $(wildcard *.c)
OBJECTS := $(patsubst %.c,%.o,$(wildcard *.c))
EMPTY := $(strip)
