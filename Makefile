VERSION := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2> /dev/null | sed 's/^v//')
ifndef VERSION
	VERSION = dev
endif
COMMIT := $(shell git rev-parse HEAD)
GO_BUILDOPTS := -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}"
GO_FILES = $(shell find . -type f -name '*.go')

.PHONY: all
all: clean exo

exo: $(GO_FILES)
	go build ${GO_BUILDOPTS} -mod=vendor -o $@

.PHONY: lint
lint: $(GO_FILES)
	golangci-lint run ./...

.PHONY: test
test: $(GO_FILES)
	go test -v -mod=vendor ./...

.PHONY: docker
docker: Dockerfile $(GO_FILES)
	docker build -f $< \
		-t exoscale/cli \
		--build-arg VERSION="${VERSION}" \
		--build-arg VCS_REF="${COMMIT}" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \
		.
	docker tag exoscale/cli:latest exoscale/cli:${VERSION}

docker-push:
	docker push exoscale/cli:latest && docker push exoscale/cli:${VERSION}

manpage:
	mkdir -p $@

.PHONY: manpages
manpages: manpage $(GO_FILES)
	go run -mod vendor doc/main.go --man-page

contrib/completion/bash:
	mkdir -p $@

.PHONY: completions
completions: contrib/completion/bash $(GO_FILES)
	go run -mod vendor completion/main.go
	mv bash_completion $</exo

.PHONY: clean
clean:
	go clean
	rm -rf exo contrib/completion manpage
