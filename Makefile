include go.mk/init.mk

GO_BIN_OUTPUT_NAME := exo

.PHONY: docker
docker:
	docker build -f $< \
		-t exoscale/cli \
		--build-arg VERSION="${VERSION}" \
		--build-arg VCS_REF="${GIT_REVISION}" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \
		.
	docker tag exoscale/cli:latest exoscale/cli:${VERSION}

docker-push:
	docker push exoscale/cli:latest && docker push exoscale/cli:${VERSION}

manpage:
	mkdir -p $@

.PHONY: manpages
manpages: manpage
	$(GO) run -mod vendor doc/main.go --man-page

.PHONY: completions
completions:
	mkdir -p contrib/completion/bash
	$(GO) run -mod vendor completion/main.go
	mv bash_completion contrib/completion/bash/exo

.PHONY: clean
clean::
	rm -rf contrib/completion manpage
