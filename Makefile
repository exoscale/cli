include go.mk/init.mk

PROJECT_URL = https://github.com/exoscale/cli

GO_BIN_OUTPUT_NAME := exo

.PHONY:
.ONESHELL:
x-cmd: ## Generates code for "exo x" experimental subsommands
	@if [ ! -f "$(shell go env GOPATH)/bin/openapi-cli-generator" ]; then
		echo "openapi-cli-generator tool not found, downloading"
		go get -u github.com/exoscale/openapi-cli-generator
	fi
	openapi-cli-generator generate -p x -n x -o cmd/internal/x/x.gen.go exoscale-v2.oas.yaml

.PHONY: docker
docker: ## Builds a Docker image containing the exo CLI
	docker build  \
		-t exoscale/cli \
		--build-arg VERSION="${VERSION}" \
		--build-arg VCS_REF="${GIT_REVISION}" \
		--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" \
		.
	docker tag exoscale/cli:latest exoscale/cli:${VERSION}

docker-push: ## Pushes the Docker image to the public Docker registry
	docker push exoscale/cli:latest && docker push exoscale/cli:${VERSION}

.PHONY: sos-certificates
sos-certificates:
	curl -sL --output sos-certs.pem https://www.exoscale.com/static/files/sos-certs.pem

.PHONY: release
release: sos-certificates
	$(MAKE) PROJECT_URL=$(PROJECT_URL) VERSION=$(VERSION) -f go.mk/public.mk $@

manpage:
	mkdir -p $@

.PHONY: manpages
manpages: manpage
	$(GO) run -mod vendor docs/main.go --man-page

.PHONY: completions
completions:
	mkdir -p contrib/completion/bash
	$(GO) run -mod vendor completion/main.go
	mv bash_completion contrib/completion/bash/exo

.PHONY: clean
clean::
	rm -rf contrib/completion manpage
