package cmd

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
)

// cliCommandImplemError represents an implementation error for a cliCommand.
type cliCommandImplemError struct {
	reason string
}

// Error returns a string representation of the cliCommandImplemError.
func (e cliCommandImplemError) Error() string {
	return fmt.Sprintf(
		"CLI command implementation error: %s. "+
			"This is a bug, and should be reported to the maintainers of this tool.",
		e.reason)
}

const CmdFlagForceHelp = "attempt to perform the operation without prompting for confirmation"

// CmdCheckRequiredFlags evaluates the specified flags as parsed in the cobra.Command flagset to check that
// their value is unset (i.e. null/empty/zero, depending on the type), and returns a multierror listing all
// flags missing a required value.
func CmdCheckRequiredFlags(cmd *cobra.Command, flags []string) error {
	var err *multierror.Error

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		for _, fn := range flags {
			if flag.Name == fn {
				var hasValue bool

				switch flag.Value.Type() {
				case "string", "stringSlice":
					if flag.Value.String() != "" {
						hasValue = true
					}

				case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64",
					"float32", "float64":
					if flag.Value.String() != "0" {
						hasValue = true
					}
				}

				if !hasValue {
					err = multierror.Append(err, fmt.Errorf("no value specified for flag %q", fn))
				}
			}
		}
	})

	return err.ErrorOrNil()
}

// CmdSetZoneFlagFromDefault attempts to set the "--zone" flag value based on the current active account's
// default zone setting if set. This is a convenience helper, there is no guarantee that the flag will be
// set once this function returns.
func CmdSetZoneFlagFromDefault(cmd *cobra.Command) {
	if cmd.Flag("zone").Value.String() == "" {
		cmd.Flag("zone").Value.Set(account.CurrentAccount.DefaultZone) // nolint:errcheck
	}
}

// CmdSetTemplateFlagFromDefault  attempts to set the "--template" flag value based on the current active account's
// default template setting if set. This is a convenience helper, there is no guarantee that the flag will be
// set once this function returns.
func CmdSetTemplateFlagFromDefault(cmd *cobra.Command) {
	if cmd.Flag("template").Value.String() == "" {
		if account.CurrentAccount.DefaultTemplate != "" {
			cmd.Flag("template").Value.Set(account.CurrentAccount.DefaultTemplate) // nolint:errcheck
		} else {
			cmd.Flag("template").Value.Set(DefaultTemplate) // nolint:errcheck
		}
	}
}

func CmdExitOnUsageError(cmd *cobra.Command, reason string) {
	cmd.PrintErrln(fmt.Sprintf("error: %s", reason))
	cmd.Usage() // nolint:errcheck
	os.Exit(1)
}

// CmdShowHelpFlags outputs flags matching the specified prefix in the provided flag set.
// This can be used for example to craft specialized usage help messages for hidden flags.
func CmdShowHelpFlags(flags *pflag.FlagSet, prefix string) {
	buf := bytes.NewBuffer(nil)
	t := table.NewEmbeddedTable(buf)

	flags.VisitAll(func(flag *pflag.Flag) {
		if strings.HasPrefix(flag.Name, prefix) {
			t.Append([]string{"--" + flag.Name, flag.Usage})
		}
	})

	t.Render()
	fmt.Print(buf)
}

// CliCommandSettings represents a CLI command settings.
type CliCommandSettings struct {
	OutputFunc func(o output.Outputter, err error) error
}

// DefaultCLICmdSettings returns a cliCommandSettings struct initialized
// with default values.
func DefaultCLICmdSettings() CliCommandSettings {
	return CliCommandSettings{
		OutputFunc: utils.PrintOutput,
	}
}

// cliCommand is the interface to implement for leveraging the automatic CLI
// command generation system based on implementer struct tags.
//
// For reference, implementers can look up the unit tests (testCLICmd struct).
// By default, all struct fields are parsed for tags: if you have private
// fields used for internal purpose, set the tag `cli:"-"` on those to exclude
// them from the CLI command evaluation process.
//
// Note: this system is an attempt at reducing the amount of boilerplate code
// required to implement CLI commands, and pagmatically supports only the
// most common CLI flag types used across the codebase (e.g. for simple CRUD
// commands). It is not one-size-fits-all and doesn't strive to be: if as a
// CLI command implementer you hit a limitation in a use case more complex
// than usual, you always have the option to use vanilla cobra/pflags, which
// is certainly easier than try to implement the missing functionnality in
// this "framework".
type cliCommand interface {
	CmdAliases() []string
	CmdShort() string
	CmdLong() string
	CmdPreRun(*cobra.Command, []string) error
	CmdRun(*cobra.Command, []string) error
}

// cliCommandFlagName returns the CLI flag name corresponding to the field
// specified from the cliCommand.
func cliCommandFlagName(c cliCommand, field interface{}) (string, error) {
	fieldValue := reflect.ValueOf(field)
	if fieldValue.Kind() != reflect.Ptr || fieldValue.IsNil() {
		return "", fmt.Errorf("field must be a non-nil pointer value")
	}

	cv := reflect.ValueOf(c).Elem()
	for i := 0; i < cv.NumField(); i++ {
		structField := cv.Type().Field(i)

		if cv.Field(i).UnsafeAddr() == fieldValue.Pointer() {
			flagName := strcase.ToKebab(structField.Name)
			if v, ok := structField.Tag.Lookup("cli-flag"); ok {
				flagName = v
			}

			return flagName, nil
		}
	}

	return "", fmt.Errorf("field not found in struct %s", cv.Type())
}

func ConvertIfSpecialEmptyMap(m map[string]string) map[string]string {
	// since it is not possible to pass an empty map
	// with a spf13/pflag https://github.com/spf13/pflag/issues/312
	// we use the special value of a map with only
	// one empty string key and the empty string value
	// as the "empty map"
	// this allows users to clear the labels of
	// a resource by passing "--label=[=]"
	v, ok := m[""]
	if ok && v == "" {
		return map[string]string{}
	}

	return m
}

func MustCLICommandFlagName(c cliCommand, field interface{}) string {
	v, err := cliCommandFlagName(c, field)
	if err != nil {
		panic(cliCommandImplemError{fmt.Sprintf("cliCommandFlagName: %s", err)})
	}

	return v
}

// cliCommandFlagSet generates a pflag.FlagSet struct from the specified
// cliCommand struct tags. Supported tags are:
//   - cli-flag:"<flag name>": override the flag name derived by default from
//     the struct field name (e.g.: cliCommand.SomeArg -> "--some-arg").
//   - cli-short:"<character>": an optional short version of the flag, e.g.
//     Zone string `cli-short:"z"` generates the CLI flag "--zone, -z".
//   - cli-usage:"<usage help>": an optional string to use as flag usage
//     help message. For positional arguments, this field is used as argument
//     label for the "use" command help.
//   - cli-hidden:"": mark the corresponding flag "hidden"
//   - cli-deprecated:"<deprecation message>": mark the corresponding flag as hidden
//     and display the deprecation message when used.
func cliCommandFlagSet(c cliCommand) (*pflag.FlagSet, error) {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	cv := reflect.ValueOf(c)

	if cv.Kind() == reflect.Ptr {
		cv = cv.Elem()
	}

	for i := 0; i < cv.NumField(); i++ {
		cTypeField := cv.Type().Field(i)

		if v, ok := cTypeField.Tag.Lookup("cli"); ok && v == "-" {
			continue
		}

		if _, ok := cTypeField.Tag.Lookup("cli-cmd"); ok {
			continue
		}

		if _, ok := cTypeField.Tag.Lookup("cli-arg"); ok {
			continue
		}

		flagName := strcase.ToKebab(cTypeField.Name)
		if v, ok := cTypeField.Tag.Lookup("cli-flag"); ok {
			flagName = v
		}

		flagShort := ""
		if v, ok := cTypeField.Tag.Lookup("cli-short"); ok {
			flagShort = v
		}

		flagUsage := ""
		if v, ok := cTypeField.Tag.Lookup("cli-usage"); ok {
			flagUsage = v
		}

		var deprecatedMsg *string
		if v, ok := cTypeField.Tag.Lookup("cli-deprecated"); ok {
			deprecatedMsg = &v
		}

		flagDefaultValue := cv.Field(i).Interface()

		switch t := cTypeField.Type.Kind(); t {
		case reflect.String:
			fs.StringP(flagName, flagShort, fmt.Sprint(flagDefaultValue), flagUsage)

		case reflect.Int64:
			fs.Int64P(flagName, flagShort, flagDefaultValue.(int64), flagUsage)

		case reflect.Bool:
			fs.BoolP(flagName, flagShort, flagDefaultValue.(bool), flagUsage)

		case reflect.Slice:
			if cTypeField.Type.Elem().Kind() != reflect.String {
				return nil, cliCommandImplemError{
					fmt.Sprintf("unsupported type []%s for field %s.%s", t, cv.Type(), cTypeField.Name),
				}
			}

			fs.StringSliceP(flagName, flagShort, flagDefaultValue.([]string), flagUsage)

		case reflect.Map:
			if cTypeField.Type.Elem().Kind() != reflect.String {
				return nil, cliCommandImplemError{
					fmt.Sprintf(
						"unsupported type map[string]%s for field %s.%s",
						t,
						cv.Type(),
						cTypeField.Name,
					),
				}
			}

			fs.StringToStringP(flagName, flagShort, flagDefaultValue.(map[string]string), flagUsage)

		default:
			return nil, cliCommandImplemError{fmt.Sprintf("unsupported type %s", t)}
		}
		if deprecatedMsg != nil {
			if err := fs.MarkDeprecated(flagName, *deprecatedMsg); err != nil {
				return nil, cliCommandImplemError{
					reason: fmt.Sprintf("unable to deprecate flag: %s, err: %s", flagName, err.Error()),
				}
			}
		}

		if _, ok := cTypeField.Tag.Lookup("cli-hidden"); ok {
			if err := fs.MarkHidden(flagName); err != nil {
				return nil, cliCommandImplemError{
					reason: fmt.Sprintf("error marking flag %q hidden: %v", flagName, err),
				}
			}
		}
	}

	return fs, nil
}

// cliCommandUse generates a string to be used as value for the cobra.Command
// "Use" field from the specified cliCommand struct tags. Supported tags are:
//   - cli-cmd:"<command name>": the name of the command (required).
//   - cli-usage:"<usage help>": an optional string to use as argument label
//     for the "use" command help.
//   - cli-arg:"<p>": declare a command line positional argument. Depending
//     on the type of the structure field (string or []string), the value of
//     <p> can either be "#" to declare a single argument which position
//     matches the one of the corresponding *ARGUMENT field* in the struct
//     type definition, or "?" to declare an optional single argument. If the
//     struct field is a []string, the result is a variadic (i.e. 0 or more)
//     list of remaining arguments; if "cli-arg:"?"` is specified, the list
//     will be marked as optional in the "use" command help.
func cliCommandUse(c cliCommand) (string, error) {
	var (
		commandName string
		use         = make([]string, 0)
	)

	cv := reflect.ValueOf(c)

	if cv.Kind() == reflect.Ptr {
		cv = cv.Elem()
	}

	for i := 0; i < cv.NumField(); i++ {
		cTypeField := cv.Type().Field(i)

		if v, ok := cv.Type().Field(i).Tag.Lookup("cli-cmd"); ok {
			commandName = v
			continue
		}

		if v, ok := cTypeField.Tag.Lookup("cli-arg"); ok {
			argLabel := strings.ToUpper(strcase.ToKebab(cv.Type().Field(i).Name))
			if u, ok := cTypeField.Tag.Lookup("cli-usage"); ok {
				argLabel = u
			}

			switch cTypeField.Type.Kind() {
			case reflect.Int64, reflect.String:
				if v == "?" {
					use = append(use, "["+argLabel+"]")
				} else {
					use = append(use, argLabel)
				}

			case reflect.Slice:
				if cTypeField.Type.Elem().Kind() != reflect.String {
					return "", cliCommandImplemError{fmt.Sprintf(
						"unsupported type []%s for field %s.%s",
						cTypeField.Type.Elem().Kind(),
						cv.Type(),
						cTypeField.Name,
					)}
				}

				if v == "?" {
					use = append(use, "["+argLabel+"]...")
				} else {
					use = append(use, argLabel+"...")
				}

			default:
				return "", cliCommandImplemError{fmt.Sprintf(
					"unsupported type %s on field %s.%s",
					cTypeField.Type.Kind(),
					cv.Type(),
					cTypeField.Name,
				)}
			}
		}
	}

	if commandName == "" {
		return "", cliCommandImplemError{
			fmt.Sprintf("`cli-cmd` tag missing from struct %s", cv.Type()),
		}
	}

	use = append([]string{commandName}, use...)

	return strings.Join(use, " "), nil
}

// CliCommandDefaultPreRun is a convenience helper function that can be used
// in cliCommand.cmdPreRun() hooks to automagically retrieve values for the
// struct flags/args fields from a cobra.Command and args provided, and set
// corresponding fields on the struct implementing the cliCommand interface.
func CliCommandDefaultPreRun(c cliCommand, cmd *cobra.Command, args []string) error { //nolint:gocyclo
	cv := reflect.ValueOf(c)

	if cv.Kind() == reflect.Ptr {
		cv = cv.Elem()
	}

	argp := 0
	for i := 0; i < cv.NumField(); i++ {
		cField := cv.Field(i)
		cTypeField := cv.Type().Field(i)

		if v, ok := cTypeField.Tag.Lookup("cli"); ok && v == "-" {
			continue
		}

		if _, ok := cTypeField.Tag.Lookup("cli-cmd"); ok {
			continue
		}

		// Positional args handling:
		if argMode, ok := cTypeField.Tag.Lookup("cli-arg"); ok {
			switch t := cTypeField.Type.Kind(); t {
			case reflect.Int64:
				if argMode == "#" {
					// Required arg
					if argp >= len(args) {
						return fmt.Errorf("missing arguments, run with --help for usage")
					}

					argVal, err := strconv.Atoi(args[argp])
					if err != nil {
						return fmt.Errorf("invalid value %q", args[argp])
					}
					cField.SetInt(int64(argVal))
				} else if argMode == "?" {
					// Optional arg
					if argp < len(args) {
						argVal, err := strconv.Atoi(args[argp])
						if err != nil {
							return fmt.Errorf("invalid value %q", args[argp])
						}
						cField.SetInt(int64(argVal))
					}
				}

			case reflect.String:
				if argMode == "#" {
					// Required arg
					if argp >= len(args) {
						return fmt.Errorf("missing arguments, run with --help for usage")
					}
					cField.SetString(args[argp])
				} else if argMode == "?" {
					// Optional arg
					if argp < len(args) {
						cField.SetString(args[argp])
					}
				}

			case reflect.Slice:
				if cTypeField.Type.Elem().Kind() != reflect.String {
					return cliCommandImplemError{fmt.Sprintf(
						"unsupported type []%s for field %s.%s",
						cTypeField.Type.Elem().Kind(),
						cv.Type(),
						cTypeField.Name,
					)}
				}

				if argp < len(args) {
					cField.Set(reflect.ValueOf(args[argp:]))
				}

			default:
				return cliCommandImplemError{fmt.Sprintf(
					"unsupported type %s on field %s.%s", t, cv.Type(), cTypeField.Name,
				)}
			}

			argp++
			continue
		}

		// Optional flags handling:
		flagName := strcase.ToKebab(cv.Type().Field(i).Name)
		if v, ok := cTypeField.Tag.Lookup("cli-flag"); ok {
			flagName = v
		}

		if cmd.Flags().Lookup(flagName) == nil {
			return cliCommandImplemError{fmt.Sprintf(
				"flag --%s not declared for field %s.%s",
				flagName,
				cv.Type(),
				cv.Type().Field(i).Name,
			)}
		}

		switch t := cTypeField.Type.Kind(); t {
		case reflect.String:
			v, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return fmt.Errorf("error retrieving value for flag --%s: %s", flagName, err)
			}
			cField.SetString(v)

		case reflect.Int64:
			v, err := cmd.Flags().GetInt64(flagName)
			if err != nil {
				return fmt.Errorf("error retrieving value for flag --%s: %s", flagName, err)
			}
			cField.SetInt(v)

		case reflect.Bool:
			v, err := cmd.Flags().GetBool(flagName)
			if err != nil {
				return fmt.Errorf("error retrieving value for flag %s: --%s", flagName, err)
			}
			cField.SetBool(v)

		case reflect.Slice:
			if cv.Type().Field(i).Type.Elem().Kind() != reflect.String {
				return cliCommandImplemError{
					fmt.Sprintf(
						"unsupported type []%s for field %s.%s",
						cv.Type().Field(i).Type.Elem().Kind(),
						cv.Type(),
						cv.Type().Field(i).Name),
				}
			}

			v, err := cmd.Flags().GetStringSlice(flagName)
			if err != nil {
				return fmt.Errorf("error retrieving value for flag %s: %s", flagName, err)
			}
			cField.Set(reflect.ValueOf(v))

		case reflect.Map:
			if cv.Type().Field(i).Type.Elem().Kind() != reflect.String {
				return cliCommandImplemError{
					fmt.Sprintf(
						"unsupported type map[string]%s for field %s.%s",
						cv.Type().Field(i).Type.Elem().Kind(),
						cv.Type(),
						cv.Type().Field(i).Name),
				}
			}

			v, err := cmd.Flags().GetStringToString(flagName)
			if err != nil {
				return fmt.Errorf("error retrieving value for flag %s: %s", flagName, err)
			}
			cField.Set(reflect.ValueOf(v))

		default:
			return cliCommandImplemError{fmt.Sprintf("unsupported type %s", t)}
		}
	}

	return nil
}

// RegisterCLICommand registers the specified cliCommand instance to the
// current CLI framework (currently Cobra).
func RegisterCLICommand(parent *cobra.Command, c cliCommand) error {
	cmdUse, err := cliCommandUse(c)
	if err != nil {
		return fmt.Errorf("error initializing CLI command: %s", err)
	}

	cmd := &cobra.Command{
		Use:     cmdUse,
		Aliases: c.CmdAliases(),
		Short:   c.CmdShort(),
		Long:    c.CmdLong(),
		PreRunE: c.CmdPreRun,
		RunE:    c.CmdRun,
	}

	cmdFlags, err := cliCommandFlagSet(c)
	if err != nil {
		return fmt.Errorf("error initializing CLI command: %s", err)
	}
	if cmdFlags != nil {
		cmdFlags.VisitAll(func(flag *pflag.Flag) {
			cmd.Flags().AddFlag(flag)
		})
	}

	parent.AddCommand(cmd)

	return nil
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "output",
		Short: "Output formatting usage",
		Long: `The exo CLI tool allows you to customize its commands output using different
formats such as table, JSON or text template using the "--output-format" flag
("-O" in short version).

By default the "table" format is applied, best suited for human reading. In
case you need to process a command output with other CLI tools, for example
in a shell script, you can either use the "json" output format (e.g. to be
piped into jq):

	$ exo config list -O json | jq .
	[
	  {
	    "name": "alice",
	    "default": true
	  },
	  {
	    "name": "bob",
	    "default": false
	  }
	]

The "text" format prints a command's output in plain text according to a
user-defined formatting template provided with the "--output-template" flag:

	$ exo config list --output-template '{{ .Name }}' | sort
	alice
	bob

The templating format is Go's text/template, which allows conditional
formatting. For example to display a "*" next to the default configuration
account:

	$ exo config list --output-template '{{ .Name }}{{ if .Default }}*{{ end }}'
	alice*
	bob

If no output template is provided, the default is to print all fields
separated by a tabulation (\t) character so the output can be parsed by a
delimiter-based processing tool such as cut(1) or AWK.

Each CLI "show"/"list" command supports specific template annotations that are
documented in the command's help page (e.g. "exo config list --help").

Note: in "list" commands the templating is applied per entry, so it is not
necessary to range on iterable data types. Each entry is terminated by a line
return character.

For the complete Go templating reference, see https://godoc.org/text/template
`,
	},
	)
}
