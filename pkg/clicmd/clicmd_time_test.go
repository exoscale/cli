package clicmd

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type testTimeCLICmd struct {
	_ bool `cli-cmd:"test-time"`

	Arg time.Time `cli-arg:"?" cli-usage:"TIMESTAMP"`

	Since time.Time `cli-short:"s" cli-usage:"start time"`
}

func (c *testTimeCLICmd) CmdAliases() []string { return nil }
func (c *testTimeCLICmd) CmdShort() string     { return "" }
func (c *testTimeCLICmd) CmdLong() string      { return "" }
func (c *testTimeCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}
func (c *testTimeCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

func Test_cliCommandFlagSet_time(t *testing.T) {
	tests := []struct {
		name         string
		defaultSince time.Time
		wantDefValue string
		wantUsage    string
	}{
		{
			name:      "no default",
			wantUsage: "  -s, --since string   start time\n",
		},
		{
			name:         "default",
			defaultSince: time.Date(2023, 6, 7, 10, 0, 0, 0, time.FixedZone("", 2*3600)),
			wantDefValue: "2023-06-07T10:00:00+02:00",
			wantUsage:    "  -s, --since string   start time (default \"2023-06-07T10:00:00+02:00\")\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := cliCommandFlagSet(&testTimeCLICmd{Since: tt.defaultSince})
			assert.NoError(t, err)

			flag := fs.Lookup("since")
			if assert.NotNil(t, flag) {
				assert.Equal(t, "string", flag.Value.Type())
				assert.Equal(t, "s", flag.Shorthand)
				assert.Equal(t, tt.wantDefValue, flag.DefValue)
			}
			assert.Equal(t, tt.wantUsage, fs.FlagUsages())
		})
	}
}

func Test_cliCommandUse_time(t *testing.T) {
	use, err := cliCommandUse(&testTimeCLICmd{})
	assert.NoError(t, err)
	assert.Equal(t, "test-time [TIMESTAMP]", use)
}

func Test_registerCobraCommand_time(t *testing.T) {
	ts := time.Date(2023, 6, 7, 10, 0, 0, 0, time.UTC)
	plus2 := time.Date(2023, 6, 7, 10, 0, 0, 0, time.FixedZone("", 2*3600))

	tests := []struct {
		name          string
		defaultSince  time.Time
		args          []string
		expectedSince time.Time
		expectedArg   time.Time
		wantErr       string
	}{
		{
			name: "unset",
			args: []string{"test-time"},
		},
		{
			name:          "default value",
			defaultSince:  ts,
			args:          []string{"test-time"},
			expectedSince: ts,
		},
		{
			name:          "long flag, UTC",
			args:          []string{"test-time", "--since", "2023-06-07T10:00:00Z"},
			expectedSince: ts,
		},
		{
			name:          "short flag, offset",
			args:          []string{"test-time", "-s", "2023-06-07T10:00:00+02:00"},
			expectedSince: plus2,
		},
		{
			name:          "fractional seconds",
			args:          []string{"test-time", "--since", "2023-06-07T10:00:00.5Z"},
			expectedSince: ts.Add(500 * time.Millisecond),
		},
		{
			name:         "empty flag value unsets the default",
			defaultSince: ts,
			args:         []string{"test-time", "--since", ""},
		},
		{
			name:        "positional arg",
			args:        []string{"test-time", "2023-06-07T10:00:00Z"},
			expectedArg: ts,
		},
		{
			name:    "date only flag value",
			args:    []string{"test-time", "--since", "2023-06-07"},
			wantErr: `invalid argument "2023-06-07" for "-s, --since" flag: parsing time "2023-06-07" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`,
		},
		{
			name:    "missing time zone",
			args:    []string{"test-time", "--since", "2023-06-07T10:00:00"},
			wantErr: `invalid argument "2023-06-07T10:00:00" for "-s, --since" flag: parsing time "2023-06-07T10:00:00" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "Z07:00"`,
		},
		{
			name:    "invalid positional arg",
			args:    []string{"test-time", "soon"},
			wantErr: `invalid value for TIMESTAMP: parsing time "soon" as "2006-01-02T15:04:05Z07:00": cannot parse "soon" as "2006"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{SilenceErrors: true, SilenceUsage: true}
			testCmd := &testTimeCLICmd{Since: tt.defaultSince}
			assert.NoError(t, RegisterCLICommand(rootCmd, testCmd))

			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			assert.Equal(t, tt.wantErr, errString(err))
			assert.True(t, tt.expectedSince.Equal(testCmd.Since), "since: want %s, got %s", tt.expectedSince, testCmd.Since)
			assert.True(t, tt.expectedArg.Equal(testCmd.Arg), "arg: want %s, got %s", tt.expectedArg, testCmd.Arg)
		})
	}
}
