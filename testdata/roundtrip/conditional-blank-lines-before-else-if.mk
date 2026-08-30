ifeq ($(CI),)
VAR := local


else ifeq ($(CI),true)
VAR := ci

endif
