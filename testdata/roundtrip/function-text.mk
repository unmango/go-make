SOURCES := foo.c bar.c
OBJECTS := $(patsubst %.c,%.o,$(SOURCES))
NAMES := $(sort $(notdir $(SOURCES)))
SPACED := $(subst a, b,text)
