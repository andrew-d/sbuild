# Get files to include in the bindata file
IGNORE_DEPS := Makefile $(shell find . -name '*.go')
DEPFILES    := $(filter-out $(IGNORE_DEPS),$(shell find . -type f))

# Figure out our directory name to use as a package name.
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURR_DIR 	  := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_PATH))))
PKGNAME       := $(shell basename $(CURR_DIR))

bindata.go: $(DEPFILES)
	go-bindata -ignore='(.*\.go|Makefile)$$' -pkg=$(PKGNAME) .
