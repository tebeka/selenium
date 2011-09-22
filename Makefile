include $(GOROOT)/src/Make.inc

TARG=selenium

GOFILES= \
	selenium.go \
	remote.go \
	selenium_rc.go

include $(GOROOT)/src/Make.pkg
