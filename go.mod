module github.com/exoscale/cli

replace gopkg.ini/ini.v1 => github.com/go-ini/ini v1.42.0

require (
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/credentials v1.1.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.0.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.2.0
	github.com/aws/smithy-go v1.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/exoscale/egoscale v0.100.3
	github.com/exoscale/openapi-cli-generator v1.1.0
	github.com/fatih/camelcase v1.0.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/iancoleman/strcase v0.2.0
	github.com/izumin5210/gentleman-logger v1.0.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/manifoldco/promptui v0.3.2
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.8.2
	github.com/vbauerster/mpb/v4 v4.12.2
	github.com/withfig/autocomplete-tools/integrations/cobra v1.2.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.1.0
	golang.org/x/term v0.5.0
	golang.org/x/text v0.7.0
	gopkg.in/h2non/gentleman.v2 v2.0.4
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/VividCortex/ewma v1.1.1 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/alecthomas/chroma v0.7.3 // indirect
	github.com/alecthomas/gometalinter v2.0.11+incompatible // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.1.1 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/client9/misspell v0.3.4 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/danielgtaylor/go-jmespath-plus v0.0.0-20200228063638-e0b6f132acba // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/lint v0.0.0-20181026193005-c67002cb31c3 // indirect
	github.com/google/shlex v0.0.0-20181106134648-c34317bd91bf // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20180909121442-1003c8bd00dc // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/izumin5210/httplogger v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a // indirect
	github.com/lunixbochs/vtclean v0.0.0-20180621232353-2d01aacdc34a // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tsenart/deadcode v0.0.0-20160724212837-210d2dc333e9 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/tools v0.1.12 // indirect
	gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20180810215634-df19058c872c // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

go 1.20
