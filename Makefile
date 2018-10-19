version := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2> /dev/null | echo 'latest')

.PHONY: exo
exo: Dockerfile
	docker build -f $^ \
		-t exoscale/${@}:${version} \
		--build-arg VERSION="${version}" \
		--build-arg VCS_REF="$(shell git rev-parse HEAD)" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \
		.
