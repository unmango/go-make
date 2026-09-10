ifeq ($(CI),)
VAR := local

else
VAR := ci
endif
