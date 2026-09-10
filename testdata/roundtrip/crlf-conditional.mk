ifeq ($(CI),true)
VAR := ci
else
VAR := local
endif

all:
	echo $(VAR)
