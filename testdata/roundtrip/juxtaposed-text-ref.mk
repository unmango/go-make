prefix$(FOO): dep$(BAR)
NAME := pre$(FOO)
BRACED := pre${FOO}
MID := a$(FOO)b
CHAIN := a$(FOO)-b$(BAR).c
FLAG := -I$(DIR)
