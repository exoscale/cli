package clicmd

const (
	DefaultEnvironment   = "api"
	DefaultTemplate      = "Linux Ubuntu 22.04 LTS 64-bit"
	DefaultSosEndpoint   = "https://sos-{zone}.exo.io"
	DefaultZone          = "ch-dk-2"
	DefaultOutputFormat  = "table"
	DefaultClientTimeout = 20 // TODO: not used, check where the client timeout is defined and used

	// Exit codes
	ExitCodeInterrupt = 130 // SIGINT (Ctrl+C) - POSIX convention: 128 + signal_number (2)
)
