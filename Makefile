version := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2> /dev/null || echo 'latest')

GO_FILES = $(shell find . -type f -name '*.go')

.PHONY: all
all: clean exo

exo: $(GO_FILES)
	go build -mod=vendor -o $@

.PHONY: lint
lint: $(GO_FILES)
	golangci-lint run ./...

.PHONY: test
test: $(GO_FILES)
	go test -v -mod=vendor ./...

.PHONY: docker
docker: Dockerfile $(GO_FILES)
	docker build -f $^ \
		-t exoscale/cli:${version} \
		--build-arg VERSION="${version}" \
		--build-arg VCS_REF="$(shell git rev-parse HEAD)" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \
		.

manpage:
	mkdir -p $@
	go run -mod vendor doc/main.go --man-page

.PHONY: manpages
manpages: manpage $(GO_FILES)
	$(foreach page,$(shell find $< -type f -iname '*.1'), gzip $(page);)

bash_completion: $(GO_FILES)
	go run -mod vendor completion/main.go

.PHONY: clean
clean:
	go clean
	rm -rf exo bash_completion manpage
