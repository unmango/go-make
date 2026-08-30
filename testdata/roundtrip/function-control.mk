LIST := a b c
OBJECTS := $(foreach var,$(LIST),$(var).o)
CHOICE := $(if $(findstring a,$(LIST)),yes,no)
