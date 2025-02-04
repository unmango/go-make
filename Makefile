_ := $(shell mkdir -p .make bin)

WORKING_DIR := $(shell pwd)
LOCALBIN    := ${WORKING_DIR}/bin

export GOBIN := ${LOCALBIN}

DEVCTL := ${LOCALBIN}/devctl
GINKGO := ${LOCALBIN}/ginkgo

ifeq ($(CI),)
TEST_FLAGS := --label-filter !E2E
else
TEST_FLAGS := --github-output --race --trace --coverprofile=cover.profile
endif

build: .make/build
test: .make/test
format: .make/go-fmt .make/dprint-fmt
tidy: go.sum
dev: .envrc bin/ginkgo bin/devctl

test_all:
	$(GINKGO) run -r ./

validate_codecov: .make/validate_codecov

cover: cover.profile
	go tool cover -func=$<

clean:
	rm -rf .make
	rm -f cover.profile

cover.profile: $(shell $(DEVCTL) list --go) | bin/ginkgo bin/devctl
	$(GINKGO) run --coverprofile=cover.profile -r ./

go.sum: go.mod $(shell $(DEVCTL) list --go) | bin/devctl
	go mod tidy

%_suite_test.go: | bin/ginkgo
	cd $(dir $@) && $(GINKGO) bootstrap

%_test.go: | bin/ginkgo
	cd $(dir $@) && $(GINKGO) generate $(notdir $*)

bin/ginkgo: go.mod
	go install github.com/onsi/ginkgo/v2/ginkgo

bin/devctl: .versions/devctl
	go install github.com/unmango/devctl/cmd@v$(shell cat $<)
	mv ${LOCALBIN}/cmd $@

bin/dprint: .versions/dprint .make/dprint/install.sh
	DPRINT_INSTALL=${WORKING_DIR} .make/dprint/install.sh

.envrc: hack/example.envrc
	cp $< $@

.make:
	mkdir -p $@

.make/build: $(shell $(DEVCTL) list --go --exclude-tests) | bin/devctl .make
	go build ./...
	@touch $@

.make/test: $(shell $(DEVCTL) list --go) | bin/ginkgo bin/devctl .make
	$(GINKGO) run ${TEST_FLAGS} $(sort $(dir $?))
	@touch $@

.make/validate_codecov: codecov.yml | .make
	curl -X POST --data-binary @codecov.yml https://codecov.io/validate
	@touch $@

.make/go-fmt: $(shell $(DEVCTL) list --go)
	go fmt
	@touch $@

# Hilariously, when the script is named `dprint-install.sh`, this line kills the install script itself
# https://github.com/dprint/dprint/blob/00e8f5e9895147b20fe70a0e4e5437bd54d928e8/website/src/assets/install.sh#L60
.make/dprint/install.sh:
	mkdir -p $(dir $@)
	curl -fsSL https://dprint.dev/install.sh -o $@
	chmod +x $@

.make/dprint-fmt: README.md | bin/dprint
	dprint fmt
	@touch $@
