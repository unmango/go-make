reverse = $(2) $(1)
A := $(call reverse,a,b)
$(eval B := $(A))
