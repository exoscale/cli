package clicmd

const (
	DefaultEnvironment        = "api"
	DefaultInstanceType       = "medium"                        // TODO: Move this variable in the packages that use it
	DefaultInstanceTypeFamily = "standard"                      // TODO: Move this variable in the packages that use it
	DefaultTemplate           = "Linux Ubuntu 22.04 LTS 64-bit" // TODO: move it outside of this package + it's heper function
	DefaultTemplateVisibility = "public"                        // TODO: move it into instance and instance pool package
	DefaultSosEndpoint        = "https://sos-{zone}.exo.io"
	DefaultZone               = "ch-dk-2" // TODO: shouldn't be used in subcommands
	DefaultOutputFormat       = "table"
	DefaultClientTimeout      = 20 // TODO: not used, check where the client timeout is defined and used

	// Exit codes
	ExitCodeInterrupt = 130 // SIGINT (Ctrl+C) - POSIX convention: 128 + signal_number (2)
)
