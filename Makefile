include go.mk/init.mk

PROJECT_URL = https://github.com/exoscale/cli

GO_BIN_OUTPUT_NAME := exo

.PHONY:
.ONESHELL:
x-cmd:
	@if [ ! -f "$(shell go env GOPATH)/bin/openapi-cli-generator" ]; then
		echo "openapi-cli-generator tool not found, downloading"
		go get -u github.com/danielgtaylor/openapi-cli-generator
	fi
	ln -s exoscale-v2.oas.yaml x.yaml
	openapi-cli-generator generate x.yaml
	rm -f x.yaml
	sed -i -re "s/^package main$$/package x/" x.go
	mv -f x.go cmd/internal/x/x.gen.go

.PHONY: docker
docker:
	docker build  \
		-t exoscale/cli \
		--build-arg VERSION="${VERSION}" \
		--build-arg VCS_REF="${GIT_REVISION}" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \
		.
	docker tag exoscale/cli:latest exoscale/cli:${VERSION}

docker-push:
	docker push exoscale/cli:latest && docker push exoscale/cli:${VERSION}

.PHONY: sos-certificates
sos-certificates:
	curl -sL --output sos-certs.pem https://www.exoscale.com/static/files/sos-certs.pem

.PHONY: release
release: sos-certificates
	$(MAKE) PROJECT_URL=$(PROJECT_URL) -f go.mk/public.mk $@

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
