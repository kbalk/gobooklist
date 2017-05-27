# Makefile for booklist.
#
# Note:  $GOPATH should be set; if not, an error will be raised.
#
# Borrowed a bit from: https://github.com/vincentbernat/hellogopher/Makefile
#

ifndef GOPATH
$(error GOPATH is not set)
endif

BINARY   = booklist
BIN      = $(GOPATH)/bin
PACKAGES = $(shell go list ./...)

default:
	go install $(PACKAGES)

##########################################################################
# Tools -- if not present, retrieve and install them.
#
GOLINT = $(BIN)/golint
$(BIN)/golint: ; $(info building golint ...)
	go get github.com/golang/lint/golint

ERRCHECK = $(BIN)/errcheck
$(BIN)/errcheck: ; $(info building errcheck ...)
	go get github.com/kisielk/errcheck

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: ; $(info building gocovmerg ...)
	go get github.com/wadey/gocovmerge

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
# Test and code coverage
#
TEST_PKGS = $(shell go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)

.PHONY: test
test:
	go test $(TEST_PKGS)

.PHONY: test-verbose
test-verbose:
	go test -v $(TEST_PKGS)

COVERAGE_PKGS    = $(shell echo $(TEST_PKGS) | tr ' ' ',')
COVERAGE_DIR     = coverage
COVERAGE_PROFILE = $(COVERAGE_DIR)/merged.profile
COVERAGE_HTML    = $(COVERAGE_DIR)/coverage.html

.PHONY: test-coverage
test-coverage: $(GOCOVMERGE)
	@mkdir -p $(COVERAGE_DIR)
	@go list -f '{{if gt (len .TestGoFiles) 0}}\
	    go test \\\
		-test.timeout=120s \\\
		-covermode count \\\
		-coverprofile $(COVERAGE_DIR)/{{.Name}}.profile \\\
		-coverpkg $(COVERAGE_PKGS) \\\
	        {{.ImportPath}}{{end}}' \
	    $(TEST_PKGS) | xargs -I {} bash -c {}
	@echo "creating merged profile ..."
	@$(GOCOVMERGE) $(COVERAGE_DIR)/*.profile > $(COVERAGE_PROFILE)
	@echo "printing function coverage ..."
	@go tool cover -func $(COVERAGE_PROFILE)
	@echo "creating html file $(COVERAGE_HTML) ..."
	@go tool cover -html $(COVERAGE_PROFILE) -o $(COVERAGE_HTML)

.PHONY: cleantest
cleantest:
	@rm -rf $(COVERAGE_DIR)

##########################################################################
# Build, install, clean
#
.PHONY: build
build:
	go build $(PACKAGES)

.PHONY: install
install:
	go install $(PACKAGES)

.PHONY: clean cleancode
clean: cleancode cleantest
cleancode:
	go clean
	@rm -f $(BIN)/$(BINARY)
