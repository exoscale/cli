package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

type testCLICmd struct {
	_ bool `cli-cmd:"test"`

	RequiredArg  string   `cli-arg:"#"`
	OptionalArgs []string `cli-arg:"?" cli-usage:"OPTION"`

	SingleString string `cli-short:"s"`
	Int64        int64  `cli-flag:"int64" cli-short:"i"`
	Bool         bool
	MultiStrings []string `cli-flag:"multi-string-value" cli-usage:"multiple strings"`
	StringsMap   map[string]string
	Deprecated   string `cli-flags:"deprecated" cli-deprecated:"use another flag instead"`

	aliases []string                                 `cli:"-"`
	short   string                                   `cli:"-"`
	long    string                                   `cli:"-"`
	preRun  func(_ *cobra.Command, _ []string) error `cli:"-"`
	run     func(_ *cobra.Command, _ []string) error `cli:"-"`
}

func (c *testCLICmd) CmdAliases() []string                              { return c.aliases }
func (c *testCLICmd) CmdShort() string                                  { return c.short }
func (c *testCLICmd) CmdLong() string                                   { return c.long }
func (c *testCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error { return c.preRun(cmd, args) }
func (c *testCLICmd) CmdRun(cmd *cobra.Command, args []string) error    { return c.run(cmd, args) }

func Test_cliCommandFlagName(t *testing.T) {
	cmd := &testCLICmd{}

	flag, err := cliCommandFlagName(cmd, &cmd.SingleString)
	require.NoError(t, err)
	require.Equal(t, "single-string", flag)

	flag, err = cliCommandFlagName(cmd, &cmd.MultiStrings)
	require.NoError(t, err)
	require.Equal(t, "multi-string-value", flag)
}

func Test_cliCommandFlagSet(t *testing.T) {
	var (
		testSingleStringValue       = "test"
		testDeprecatedValue         = "deprecated"
		testInt64Value        int64 = 42
		testBoolValue               = true
		testMultiStringsValue       = []string{"a", "b", "c"}
		testStringsMap              = map[string]string{"k1": "v1", "k2": "v2"}
	)

	cmd := &testCLICmd{
		SingleString: testSingleStringValue,
		Int64:        testInt64Value,
		Bool:         testBoolValue,
		MultiStrings: testMultiStringsValue,
		StringsMap:   testStringsMap,
		Deprecated:   testDeprecatedValue,
	}

	expected := pflag.NewFlagSet("", pflag.ExitOnError)
	expected.StringP("single-string", "s", testSingleStringValue, "")
	expected.StringP("deprecated", "", testDeprecatedValue, "")
	expected.Int64P("int64", "i", testInt64Value, "")
	expected.BoolP("bool", "", testBoolValue, "")
	expected.StringSliceP("multi-string-value", "", testMultiStringsValue, "multiple strings")
	expected.StringToStringP("strings-map", "", testStringsMap, "")
	require.NoError(t, expected.MarkDeprecated("deprecated", "use another flag instead"))

	actual, err := cliCommandFlagSet(cmd)
	require.NoError(t, err)
	// StringToString typed flags are not sorted when represented as strings, therefore we
	// cannot simply test the resulting flagSet using `require.Equal(t, expected, actual)`
	// as the test randomly fails if the expected and actual values are not in the same order.
	expected.VisitAll(func(expectedFlag *pflag.Flag) {
		if expectedFlag.Value.Type() == "stringToString" {
			actualStringsMap, err := actual.GetStringToString(expectedFlag.Name)
			require.NoError(t, err)
			require.Equal(t, actualStringsMap, testStringsMap)
			return
		}

		require.Equal(t, actual.Lookup(expectedFlag.Name), expectedFlag)
	})
}

func Test_cliCommandUse(t *testing.T) {
	cmd := &testCLICmd{}

	expected := "test REQUIRED-ARG [OPTION]..."

	actual, err := cliCommandUse(cmd)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func Test_cliCommandDefaultPreRun(t *testing.T) {
	var (
		testRequiredArg             = "required-arg"
		testDeprecatedValue         = "deprecated"
		testOptionalArgs            = []string{"optional-arg1", "optional-arg2"}
		testSingleStringValue       = "test"
		testInt64Value        int64 = 42
		testBoolValue               = true
		testMultiStringsValue       = []string{"a", "b", "c"}
		testStringsMap              = map[string]string{"k1": "v1", "k2": "v2"}
	)

	testFlags := pflag.NewFlagSet("", pflag.ExitOnError)
	testFlags.StringP("single-string", "s", "", "")
	testFlags.Int64P("int64", "i", 0, "")
	testFlags.BoolP("bool", "", false, "")
	testFlags.StringSliceP("multi-string-value", "", nil, "multiple strings")
	testFlags.StringToStringP("strings-map", "", nil, "")
	testFlags.StringP("deprecated", "", "", "")

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	tests := []struct {
		name     string
		args     args
		expected *testCLICmd
		wantErr  bool
	}{
		{
			name: "error required arg missing",
			args: args{
				cmd: func() *cobra.Command {
					testCmd := new(cobra.Command)
					testFlags.VisitAll(func(flag *pflag.Flag) { testCmd.Flags().AddFlag(flag) })
					return testCmd
				}(),
				args: []string{},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "ok without optional args",
			args: args{
				cmd: func() *cobra.Command {
					testCmd := new(cobra.Command)
					testFlags.VisitAll(func(flag *pflag.Flag) { testCmd.Flags().AddFlag(flag) })
					return testCmd
				}(),
				args: []string{testRequiredArg},
			},
			expected: &testCLICmd{
				RequiredArg:  testRequiredArg,
				MultiStrings: []string{},
				StringsMap:   map[string]string{},
			},
		},
		{
			name: "ok with optional args",
			args: args{
				cmd: func() *cobra.Command {
					testCmd := new(cobra.Command)
					testFlags.VisitAll(func(flag *pflag.Flag) { testCmd.Flags().AddFlag(flag) })
					return testCmd
				}(),
				args: append([]string{testRequiredArg}, testOptionalArgs...),
			},
			expected: &testCLICmd{
				RequiredArg:  testRequiredArg,
				OptionalArgs: testOptionalArgs,
				MultiStrings: []string{},
				StringsMap:   map[string]string{},
			},
		},
		{
			name: "ok with flags",
			args: args{
				cmd: func() *cobra.Command {
					flags := pflag.NewFlagSet("", pflag.ExitOnError)
					flags.StringP("single-string", "s", testSingleStringValue, "")
					flags.Int64P("int64", "i", testInt64Value, "")
					flags.BoolP("bool", "", testBoolValue, "")
					flags.StringSliceP("multi-string-value", "", testMultiStringsValue, "")
					flags.StringToStringP("strings-map", "", testStringsMap, "")
					flags.StringP("deprecated", "", testDeprecatedValue, "")

					testCmd := new(cobra.Command)
					flags.VisitAll(func(flag *pflag.Flag) { testCmd.Flags().AddFlag(flag) })
					return testCmd
				}(),
				args: []string{testRequiredArg},
			},
			expected: &testCLICmd{
				RequiredArg:  testRequiredArg,
				SingleString: testSingleStringValue,
				Deprecated:   testDeprecatedValue,
				Int64:        testInt64Value,
				Bool:         testBoolValue,
				MultiStrings: testMultiStringsValue,
				StringsMap:   testStringsMap,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := new(testCLICmd)
			err := CliCommandDefaultPreRun(actual, tt.args.cmd, tt.args.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}

func Test_registerCobraCommand(t *testing.T) {
	var (
		testCmdAliases = []string{"t"}
		testCmdShort   = "short"
		testCmdLong    = "long"

		testRequiredArg             = "required-arg"
		testOptionalArgs            = []string{"optional-arg1", "optional-arg2"}
		testSingleStringValue       = "test"
		testInt64Value        int64 = 42
		testBoolValue               = true
		testMultiStringsValue       = []string{"a", "b", "c"}
		testStringsMap              = map[string]string{"k1": "v1", "k2": "v2"}

		testCmdPreRunOK bool
		testCmdRunOK    bool
	)

	rootCmd := &cobra.Command{}

	testCmd := &testCLICmd{
		aliases: testCmdAliases,
		short:   testCmdShort,
		long:    testCmdLong,
	}

	testCmd.preRun = func(cmd *cobra.Command, args []string) error {
		if err := CliCommandDefaultPreRun(testCmd, cmd, args); err != nil {
			return err
		}

		testCmdPreRunOK = true
		return nil
	}

	testCmd.run = func(cmd *cobra.Command, args []string) error {
		require.Equal(t, testRequiredArg, testCmd.RequiredArg)
		require.Equal(t, testOptionalArgs, testCmd.OptionalArgs)
		require.Equal(t, testSingleStringValue, testCmd.SingleString)
		require.Equal(t, testInt64Value, testCmd.Int64)
		require.Equal(t, testBoolValue, testCmd.Bool)
		require.Equal(t, testMultiStringsValue, testCmd.MultiStrings)
		require.Equal(t, testStringsMap, testCmd.StringsMap)

		testCmdRunOK = true
		return nil
	}

	err := RegisterCLICommand(rootCmd, testCmd)
	require.NoError(t, err)
	require.Len(t, rootCmd.Commands(), 1)

	actual := rootCmd.Commands()[0]
	require.Equal(t, testCmdAliases, actual.Aliases)
	require.Equal(t, testCmdShort, actual.Short)
	require.Equal(t, testCmdLong, actual.Long)

	rootCmd.SetArgs(append([]string{
		"test",
		"--single-string=" + testSingleStringValue,
		"--int64=" + fmt.Sprint(testInt64Value),
		"--bool=" + fmt.Sprint(testBoolValue),
		"--multi-string-value=a",
		"--multi-string-value=b",
		"--multi-string-value=c",
		"--strings-map=k1=v1",
		"--strings-map=k2=v2",
		testRequiredArg,
	},
		testOptionalArgs...,
	))
	require.NoError(t, rootCmd.Execute())
	require.True(t, testCmdPreRunOK)
	require.True(t, testCmdRunOK)
}

// testValueCLICmd exercises fields whose type implements pflag.Value.
type testValueCLICmd struct {
	_ bool `cli-cmd:"test-value"`

	RequiredValue BinaryValue `cli-arg:"#"`
	OptionalValue BinaryValue `cli-arg:"?" cli-usage:"OPTIONAL"`

	FlagValue BinaryValue `cli-short:"b" cli-usage:"binary flag"`

	preRun func(_ *cobra.Command, _ []string) error `cli:"-"`
	run    func(_ *cobra.Command, _ []string) error `cli:"-"`
}

func (c *testValueCLICmd) CmdAliases() []string { return nil }
func (c *testValueCLICmd) CmdShort() string     { return "" }
func (c *testValueCLICmd) CmdLong() string      { return "" }
func (c *testValueCLICmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return c.preRun(cmd, args)
}
func (c *testValueCLICmd) CmdRun(cmd *cobra.Command, args []string) error { return c.run(cmd, args) }

func Test_cliCommandFlagSet_pflagValue(t *testing.T) {
	cmd := &testValueCLICmd{FlagValue: BinaryValue("default")}

	actual, err := cliCommandFlagSet(cmd)
	require.NoError(t, err)

	expected := &pflag.Flag{
		Name:      "flag-value",
		Shorthand: "b",
		Usage:     "binary flag",
		Value:     new(BinaryValue("default")),
		DefValue:  "ZGVmYXVsdA==",
	}
	require.Equal(t, expected, actual.Lookup("flag-value"))
}

func Test_cliCommandUse_pflagValue(t *testing.T) {
	actual, err := cliCommandUse(&testValueCLICmd{})
	require.NoError(t, err)
	require.Equal(t, "test-value REQUIRED-VALUE [OPTIONAL]", actual)
}

func Test_cliCommandDefaultPreRun_pflagValue(t *testing.T) {
	newTestCmd := func(flagValue string) *cobra.Command {
		fs, err := cliCommandFlagSet(&testValueCLICmd{})
		require.NoError(t, err)
		if flagValue != "" {
			require.NoError(t, fs.Set("flag-value", flagValue))
		}

		testCmd := new(cobra.Command)
		fs.VisitAll(func(flag *pflag.Flag) { testCmd.Flags().AddFlag(flag) })
		return testCmd
	}

	tests := []struct {
		name      string
		flagValue string
		args      []string
		expected  *testValueCLICmd
		wantErr   bool
	}{
		{
			name:    "error required arg missing",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "error invalid arg value",
			args:    []string{"not base64!"},
			wantErr: true,
		},
		{
			name:     "ok without optional arg",
			args:     []string{"AP8Q"},
			expected: &testValueCLICmd{RequiredValue: BinaryValue{0x00, 0xff, 0x10}},
		},
		{
			name: "ok with optional arg and flag",
			args: []string{"AP8Q", "aGk="},
			// Set on the flagset as cobra would do when parsing "--flag-value".
			flagValue: "ZGVmYXVsdA==",
			expected: &testValueCLICmd{
				RequiredValue: BinaryValue{0x00, 0xff, 0x10},
				OptionalValue: BinaryValue("hi"),
				FlagValue:     BinaryValue("default"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := new(testValueCLICmd)
			err := CliCommandDefaultPreRun(actual, newTestCmd(tt.flagValue), tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_cliCommandDefaultPreRun_pflagValueTypeMismatch(t *testing.T) {
	testCmd := new(cobra.Command)
	testCmd.Flags().String("flag-value", "", "")

	err := CliCommandDefaultPreRun(new(testValueCLICmd), testCmd, []string{"AP8Q"})
	require.ErrorAs(t, err, new(cliCommandImplemError))
}

func Test_registerCobraCommand_pflagValue(t *testing.T) {
	raw := []byte{0x00, 0xff, 0x10, 'h', 'i', '\n'}
	binFile := filepath.Join(t.TempDir(), "data.bin")
	require.NoError(t, os.WriteFile(binFile, raw, 0o600))

	var testCmdRunOK bool

	rootCmd := &cobra.Command{}
	testCmd := &testValueCLICmd{}
	testCmd.preRun = func(cmd *cobra.Command, args []string) error {
		return CliCommandDefaultPreRun(testCmd, cmd, args)
	}
	testCmd.run = func(_ *cobra.Command, _ []string) error {
		require.Equal(t, BinaryValue{0x00, 0xff, 0x10}, testCmd.RequiredValue)
		require.Equal(t, BinaryValue(raw), testCmd.OptionalValue)
		require.Equal(t, BinaryValue("hi"), testCmd.FlagValue)

		testCmdRunOK = true
		return nil
	}

	require.NoError(t, RegisterCLICommand(rootCmd, testCmd))

	rootCmd.SetArgs([]string{"test-value", "-b", "aGk=", "AP8Q", "fileb://" + binFile})
	require.NoError(t, rootCmd.Execute())
	require.True(t, testCmdRunOK)
}
