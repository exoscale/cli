module github.com/exoscale/cli

replace gopkg.ini/ini.v1 => github.com/go-ini/ini v1.42.0

require (
	github.com/alecthomas/chroma v0.7.3 // indirect
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/credentials v1.1.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.0.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.2.0
	github.com/aws/smithy-go v1.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/exoscale/egoscale v0.100.1
	github.com/exoscale/openapi-cli-generator v1.1.0
	github.com/fatih/camelcase v1.0.0
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/iancoleman/strcase v0.2.0
	github.com/izumin5210/gentleman-logger v1.0.0
	github.com/izumin5210/httplogger v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/manifoldco/promptui v0.3.2
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/minio/minio-go/v6 v6.0.57
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.8.1
	github.com/vbauerster/mpb/v4 v4.12.2
	github.com/withfig/autocomplete-tools/integrations/cobra v1.2.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20180810215634-df19058c872c // indirect
	gopkg.in/h2non/gentleman.v2 v2.0.4
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

go 1.16
