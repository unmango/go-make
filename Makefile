_ := $(shell mkdir -p .make bin)

WORKING_DIR := $(shell pwd)
LOCALBIN    := ${WORKING_DIR}/bin

export GOBIN := ${LOCALBIN}

GINKGO := go tool ginkgo

# Every Go source file in the repo, relative to the repo root.
GO_SOURCES := $(shell find . \( -path ./.git -o -path ./.direnv -o -path ./bin \) -prune -o -name '*.go' -print | sed 's|^\./||')
# The same list without test files.
GO_NON_TEST_SOURCES := $(filter-out %_test.go,$(GO_SOURCES))

ifeq ($(strip $(GO_SOURCES)),)
$(error no Go sources found, source discovery is broken and prerequisites would be empty)
endif

DPRINT_VERSION := $(shell cat .versions/dprint)

ifeq ($(strip $(DPRINT_VERSION)),)
$(error could not read the pinned dprint version from .versions/dprint)
endif

ifeq ($(CI),)
TEST_FLAGS := --label-filter !E2E
else
TEST_FLAGS := --github-output --race --trace --coverprofile=cover.profile
endif

build: .make/build
test: .make/test
format: .make/go-fmt .make/dprint-fmt
tidy: go.sum
dev: .envrc

.PHONY: build test test_all format tidy dev cover clean validate_codecov sync-quickref

test_all:
	$(GINKGO) run -r ./

sync-quickref:
	nix run .#sync-quickref

validate_codecov: .make/validate_codecov

cover: cover.profile
	go tool cover -func=$<

clean:
	rm -rf .make
	rm -f cover.profile

cover.profile: $(GO_SOURCES)
	$(GINKGO) run --coverprofile=cover.profile -r ./

go.sum: go.mod $(GO_SOURCES)
	go mod tidy

%_suite_test.go:
	cd $(dir $@) && $(GINKGO) bootstrap

%_test.go:
	cd $(dir $@) && $(GINKGO) generate $(notdir $*)

bin/dprint: .versions/dprint | .make/dprint/install.sh
	DPRINT_INSTALL=${WORKING_DIR} .make/dprint/install.sh ${DPRINT_VERSION}
	@touch $@

.envrc: hack/example.envrc
	cp $< $@

.make:
	mkdir -p $@

.make/build: $(GO_NON_TEST_SOURCES) | .make
	go build ./...
	@touch $@

.make/test: $(GO_SOURCES) $(wildcard testdata/*) | .make
	$(GINKGO) run ${TEST_FLAGS} $(sort $(dir $(filter-out testdata/%,$?)))
	@touch $@

.make/validate_codecov: codecov.yml | .make
	curl -X POST --data-binary @codecov.yml https://codecov.io/validate
	@touch $@

.make/go-fmt: $(GO_SOURCES)
	go fmt ./...
	@touch $@

# Hilariously, when the script is named `dprint-install.sh`, this line kills the install script itself
# https://github.com/dprint/dprint/blob/00e8f5e9895147b20fe70a0e4e5437bd54d928e8/website/src/assets/install.sh#L60
.make/dprint/install.sh:
	mkdir -p $(dir $@)
	curl -fsSL https://dprint.dev/install.sh -o $@
	chmod +x $@

.make/dprint-fmt: README.md | bin/dprint
	${LOCALBIN}/dprint fmt
	@touch $@
