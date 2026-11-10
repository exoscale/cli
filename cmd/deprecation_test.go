package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type testDeprecatedCLICmd struct {
	_ bool `cli-cmd:"leaf"`

	preRunOK bool `cli:"-"`
}

func (c *testDeprecatedCLICmd) CmdAliases() []string { return nil }
func (c *testDeprecatedCLICmd) CmdShort() string     { return "Leaf" }
func (c *testDeprecatedCLICmd) CmdLong() string      { return "Leaf command." }
func (c *testDeprecatedCLICmd) CmdDeprecated() string {
	return "leaf is deprecated, use other instead"
}
func (c *testDeprecatedCLICmd) CmdPreRun(_ *cobra.Command, _ []string) error {
	c.preRunOK = true
	return nil
}
func (c *testDeprecatedCLICmd) CmdRun(_ *cobra.Command, _ []string) error { return nil }

func TestApplyDeprecations(t *testing.T) {
	const groupMessage = "group is deprecated, use other instead"

	var childPreRunOK bool

	root := &cobra.Command{Use: "exo"}
	other := &cobra.Command{Use: "other", Short: "Other", Run: func(*cobra.Command, []string) {}}
	group := Deprecate(&cobra.Command{Use: "group", Short: "Group"}, groupMessage)
	// Added after the group is marked, like commands registered by later init()s.
	child := &cobra.Command{
		Use:    "child",
		Short:  "Child",
		Long:   "Child command.",
		PreRun: func(*cobra.Command, []string) { childPreRunOK = true },
		Run:    func(*cobra.Command, []string) {},
	}
	root.AddCommand(other, group)
	group.AddCommand(child)

	leaf := &testDeprecatedCLICmd{}
	require.NoError(t, RegisterCLICommand(root, leaf))

	ApplyDeprecations(root)
	ApplyDeprecations(root) // must not apply the changes twice

	require.Equal(t, "Group (deprecated)", group.Short)
	require.Equal(t, "DEPRECATED: "+groupMessage, group.Long)
	require.Equal(t, "Child (deprecated)", child.Short)
	require.Equal(t, "DEPRECATED: "+groupMessage+"\n\nChild command.", child.Long)
	require.Equal(t, "Other", other.Short)
	require.Empty(t, other.Long)

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	t.Run("deprecated commands stay listed in the help", func(t *testing.T) {
		stdout.Reset()
		root.SetArgs([]string{"--help"})
		require.NoError(t, root.Execute())
		require.Contains(t, stdout.String(), "Group (deprecated)")
		require.Contains(t, stdout.String(), "Leaf (deprecated)")
	})

	t.Run("a group's subcommand warns on stderr", func(t *testing.T) {
		stdout.Reset()
		stderr.Reset()
		root.SetArgs([]string{"group", "child"})
		require.NoError(t, root.Execute())
		require.Equal(t, "warning: "+groupMessage+"\n", stderr.String())
		require.Empty(t, stdout.String())
		require.True(t, childPreRunOK, "the original PreRun must still run")
	})

	t.Run("a cliCommand with CmdDeprecated warns on stderr", func(t *testing.T) {
		stderr.Reset()
		root.SetArgs([]string{"leaf"})
		require.NoError(t, root.Execute())
		require.Equal(t, "warning: leaf is deprecated, use other instead\n", stderr.String())
		require.True(t, leaf.preRunOK, "the original PreRunE must still run")
	})

	t.Run("other commands don't warn", func(t *testing.T) {
		stderr.Reset()
		root.SetArgs([]string{"other"})
		require.NoError(t, root.Execute())
		require.Empty(t, stderr.String())
	})
}
