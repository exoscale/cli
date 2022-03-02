include go.mk/init.mk
include go.mk/public.mk

PACKAGE := github.com/exoscale/egoscale

PROJECT_URL := https://$(PACKAGE)

GO_TEST_EXTRA_ARGS := -mod=readonly -v

GOLANGCI_EXTRA_ARGS := --modules-download-mode=readonly

.PHONY: oapigen
oapigen:
	cd v2/oapi/
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.1
	wget -q https://openapi-v2.exoscale.com/source.json
	go generate
	@rm source.json
