target:
ifdef VERBOSE
ifeq ($(CI),1)
	echo both
endif
	echo outer
endif
