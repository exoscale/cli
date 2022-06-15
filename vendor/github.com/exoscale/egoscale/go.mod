module github.com/exoscale/egoscale

require (
	github.com/deepmap/oapi-codegen v1.9.1
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

go 1.17

retract v1.19.0 // Published accidentally.
