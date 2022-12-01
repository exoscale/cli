module github.com/exoscale/egoscale

require (
	github.com/deepmap/oapi-codegen v1.9.1
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

go 1.17

retract v1.19.0 // Published accidentally.
