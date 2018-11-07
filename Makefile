version := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2> /dev/null || echo 'latest')

all: exo

exo:
	go build -mod=vendor -o $@

lint:
	golangci-lint run ./...

test:
	go test -v -mod=vendor ./...

docker: Dockerfile
	docker build -f $^ \
		-t exoscale/cli:${version} \
		--build-arg VERSION="${version}" \
		--build-arg VCS_REF="$(shell git rev-parse HEAD)" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \

clean:
	go clean
	rm -f exo

.PHONY: docker all exo lint test clean
