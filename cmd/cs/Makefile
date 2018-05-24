PKG=github.com/exoscale/egoscale

GOOS?=linux
GOARCH?=amd64

GOPATH=$(CURDIR)/.gopath
DEP=$(GOPATH)/bin/dep

SRCS=$(wildcard *.go)
OSES=windows darwin linux
ifeq ($(GOOS),windows)
	SUFFIX=.exe
endif
ifeq ($(GOOS),darwin)
	SUFFIX=-darwin
endif
ifeq ($(GOOS),linux)
	SUFFIX=
endif

DEST=build
DEBUG_BIN=cs
BIN=$(DEST)/$(DEBUG_BIN)$(SUFFIX)

export GOPATH

.PHONY: default
default: $(DEBUG_BIN)

$(GOPATH)/src/$(PKG):
	mkdir -p $(GOPATH)
	go get -u github.com/golang/dep/cmd/dep
	mkdir -p $(shell dirname $(GOPATH)/src/$(PKG))
	ln -sf ../../../../../.. $(GOPATH)/src/$(PKG)

$(DEBUG_BIN): $(SRCS)
	(cd $(GOPATH)/src/$(PKG)/cmd/cs && \
		go build \
			-o $@ \
			.)

$(BIN): $(SRCS)
	(cd $(GOPATH)/src/$(PKG)/cmd/cs && \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-ldflags "-s" \
			-o $@ \
			.)

.PHONY: deps
deps: $(GOPATH)/src/$(PKG)
	(cd $(GOPATH)/src/$(PKG) && \
		$(DEP) ensure -v -vendor-only)
	(cd $(GOPATH)/src/$(PKG)/cmd/cs && \
		$(DEP) ensure -v)

.PHONY: deps-update
deps-update: deps
	(cd $(GOPATH)/src/$(PKG)/cmd/cs && \
		$(DEP) ensure -v -update)

.PHONY: signature
signature: $(BIN).asc

$(BIN).asc: $(BIN)
	rm -f $(BIN).asc
	gpg -a -u ops@exoscale.ch --output $@ --detach-sign $<

.PHONY: release
release: deps
	$(foreach goos, $(OSES), \
		GOOS=$(goos) $(MAKE) signature;)

.PHONY: clean
clean:
	$(RM) -r $(DEST)
	go clean
