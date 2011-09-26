include $(GOROOT)/src/Make.inc

TARG=selenium

GOFILES= \
	selenium.go \
	remote.go \
	common.go 

include $(GOROOT)/src/Make.pkg
