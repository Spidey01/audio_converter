
# The default install prefix.
PREFIX ?= /usr/local

# Use V=1 or whatever to run commands in verbose mode where useful.
ifdef V
	verbose = -v
else
	verbose =
endif


all: build check

build:
	GOBIN=$(CURDIR)/bin go install ./...

install:
	GOBIN=$(DESTDIR)$(PREFIX)/bin go install ./...

check:
	go test $(verbose) ./...

clean:
	-rm $(verbose) -rf bin

.PHONY: all build install check clean