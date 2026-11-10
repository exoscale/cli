package clicmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	deprecatedAnnotation         = "exo:deprecated"
	deprecationAppliedAnnotation = "exo:deprecation-applied"
)

// cliCommandDeprecated is an optional interface a cliCommand can implement to
// mark itself as deprecated: a non-empty CmdDeprecated() message has the same
// effect as calling Deprecate on the registered command.
type cliCommandDeprecated interface {
	CmdDeprecated() string
}

// Deprecate marks cmd and every command under it as deprecated, and returns
// cmd so it can be used in a command declaration.
//
// The message says what is deprecated and what to use instead, for example:
//
//	Private Networks are deprecated, use VPCs instead (see "exo networking vpc --help")
//
// Unlike cobra's Command.Deprecated, deprecated commands stay listed in the
// help: ApplyDeprecations appends "(deprecated)" to their short description,
// prepends the message to their long description, and makes them print it as
// a warning on stderr when they run.
func Deprecate(cmd *cobra.Command, message string) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[deprecatedAnnotation] = message

	return cmd
}

// ApplyDeprecations updates the help and run hooks of the commands marked with
// Deprecate under root, including their subcommands. It must be called once the
// whole command tree is built, before executing it or generating its docs.
// Calling it again doesn't change commands already updated.
func ApplyDeprecations(root *cobra.Command) {
	applyDeprecations(root, "")
}

func applyDeprecations(cmd *cobra.Command, inherited string) {
	message := inherited
	if m, ok := cmd.Annotations[deprecatedAnnotation]; ok && m != "" {
		message = m
	}

	if message != "" {
		applyDeprecation(cmd, message)
	}

	for _, sub := range cmd.Commands() {
		applyDeprecations(sub, message)
	}
}

func applyDeprecation(cmd *cobra.Command, message string) {
	if _, ok := cmd.Annotations[deprecationAppliedAnnotation]; ok {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[deprecationAppliedAnnotation] = ""

	notice := "DEPRECATED: " + message

	cmd.Short += " (deprecated)"
	if cmd.Long == "" {
		cmd.Long = notice
	} else {
		cmd.Long = notice + "\n\n" + cmd.Long
	}

	if !cmd.Runnable() {
		return
	}

	preRunE, preRun := cmd.PreRunE, cmd.PreRun
	cmd.PreRun = nil
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		_, _ = fmt.Fprintf(c.ErrOrStderr(), "warning: %s\n", message)

		switch {
		case preRunE != nil:
			return preRunE(c, args)
		case preRun != nil:
			preRun(c, args)
		}

		return nil
	}
}
