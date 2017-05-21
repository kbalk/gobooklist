# $GOPATH should be set; if not, an error will be raised.
#
# Borrowed a bit from: https://github.com/vincentbernat/hellogopher/Makefile
#

ifndef GOPATH
$(error GOPATH is not set)
endif

BINARY  = booklist
BIN     = $(GOPATH)/bin
PACKAGES = $(shell go list ./...)

##########################################################################
# Tools
#
GOLINT = $(BIN)/golint
$(BIN)/golint: ; $(info building golint ...)
	go get github.com/golang/lint/golint

ERRCHECK = $(BIN)/errcheck
$(BIN)/errcheck: ; $(info building errcheck ...)
	go get github.com/kisielk/errcheck

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: ; $(info building gocovmerg ...e)
	go get github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: ; $(info building gocov ...)
	go get github.com/axw/gocov/...

##########################################################################
# Code formatting, inspection
#
.PHONY: fmt
fmt: ; $(info running go fmt ...)
	@ret=0 && for d in $$(go list -f '{{.Dir}}' ./...); do \
		(cd $$d; go fmt *.go) || ret=$$? ; \
	done ; exit $$ret

.PHONY: vet
vet: ; $(info running go vet ...)
	@ret=0 && for d in $$(go list -f '{{.Dir}}' ./...); do \
		(cd $$d; go vet *.go) || ret=$$? ; \
	done ; exit $$ret

.PHONY: lint
lint: $(GOLINT) ; $(info running golint ...)
	@$(GOLINT) -set_exit_status $(PACKAGES)

.PHONY: errcheck
errcheck: $(ERRCHECK) ; $(info running errcheck ...)
	@$(ERRCHECK) -exclude errcheck_excludes.txt $(PACKAGES)

.PHONY: inspect
inspect: fmt vet lint errcheck

##########################################################################
# Tests
#

##########################################################################
# Build, install, clean
#
.PHONY: build
build:
	go build $(PACKAGES)

.PHONY: install
install:
	go install $(PACKAGES)

.PHONY: clean
clean:
	go clean
	@rm -f $(BIN)/$(BINARY)

default:
	go install
