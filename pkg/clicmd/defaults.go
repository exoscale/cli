package clicmd

const (
	DefaultEnvironment  = "api"
	DefaultTemplate     = "Linux Ubuntu 22.04 LTS 64-bit"
	DefaultSosEndpoint  = "https://sos-{zone}.exo.io"
	DefaultZone         = "ch-dk-2"
	DefaultOutputFormat = "table"

	// Exit codes
	ExitCodeInterrupt = 130 // SIGINT (Ctrl+C) - POSIX convention: 128 + signal_number (2)
)
