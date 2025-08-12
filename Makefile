
# The default install prefix.
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
DOCDIR ?= $(PREFIX)/share/audio_converter

# Use V=1 or whatever to run commands in verbose mode where useful.
ifdef V
	verbose = -v
else
	verbose =
endif
install := install $(verbose)

all: build

help:
	@echo Set PREFIX to the install location. Default is '$(PREFIX)'
	@echo The usual DESTDIR is supported for packaging PREFIX.
	@echo BINDIR and DOCDIR can be overridden separately from prefix.
	@echo Use build, install, check, and clean for the obvious things.
	@echo Set V for some verbosity.

build:
	GOBIN=$(CURDIR)/build go install ./...

install: build
	$(install) -d $(DESTDIR)$(BINDIR)
	$(install) build/* $(DESTDIR)$(BINDIR)
	$(install) -d $(DESTDIR)$(DOCDIR)
	$(install) LICENSE.txt NOTICE README.md $(DESTDIR)$(DOCDIR)

check:
	go test $(verbose) ./...

clean:
	-rm $(verbose) -rf bin

.PHONY: all build install check clean