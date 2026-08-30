target:
ifdef VERBOSE
	echo loud
else ifeq ($(CI),1)
	echo ci
else
	echo quiet
endif
