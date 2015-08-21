SUBMAKEFILES := $(shell find . -mindepth 2 -name 'Makefile')
GO_FILES 	 := $(shell find . -name '*.go')

ifeq ($(V),1)
Q :=
else
Q := @
endif

all: sbuild

sbuild: $(GO_FILES) godeps
	$(Q)go build -v -o $@ ./cmd/sbuild

.PHONY: godeps
godeps:
	$(Q)for dir in $(SUBMAKEFILES); do \
		$(MAKE) -C $$(dirname $$dir); \
	done
