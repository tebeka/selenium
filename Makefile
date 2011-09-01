include $(GOROOT)/src/Make.inc

TARG=selenium

GOFILES= \
	selenium.go \
	commands.go

include $(GOROOT)/src/Make.pkg
