include $(GOROOT)/src/Make.inc

TARG=selenium

GOFILES= \
	selenium.go \
	remote.go

include $(GOROOT)/src/Make.pkg
