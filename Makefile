include go.mk/init.mk
include go.mk/public.mk

PROJECT_URL = https://github.com/exoscale/cli
GO_BIN_OUTPUT_NAME := exo
OAS_FILE := public-api.json
RM = rm -rf

$(OAS_FILE):
	wget -O public-api.json -q https://openapi-v2.exoscale.com/source.json

.PHONY:
.ONESHELL:
x-cmd: $(OAS_FILE) ## Generates code for "exo x" experimental subcommands
	@if [ ! -f "$(shell go env GOPATH)/bin/openapi-cli-generator" ]; then
		echo "openapi-cli-generator tool not found, downloading"
		go install github.com/exoscale/openapi-cli-generator@latest
	fi
	wget -q https://openapi-v2.exoscale.com/source.json
	openapi-cli-generator generate -p x -n x -o cmd/internal/x/x.gen.go $(OAS_FILE)

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
	mkdir -p contrib/completion/bash \
		contrib/completion/powershell \
		contrib/completion/zsh
	$(GO) run -mod vendor completion/main.go bash ; mv bash_completion contrib/completion/bash/exo
	$(GO) run -mod vendor completion/main.go powershell ; mv powershell_completion contrib/completion/powershell/exo
	$(GO) run -mod vendor completion/main.go zsh ; mv zsh_completion contrib/completion/zsh/_exo

.PHONY: clean
clean::
	$(RM) contrib/completion manpage $(OAS_FILE)
