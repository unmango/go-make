define CONDITIONAL
ifeq ($(CI),)
local
else
ci
endif
endef
