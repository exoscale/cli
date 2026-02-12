package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"

	"github.com/exoscale/cli/cmd"
	_ "github.com/exoscale/cli/cmd/subcommands"
)

func main() {

	var flagError pflag.ErrorHandling
	docCmd := pflag.NewFlagSet("", flagError)
	var manPage = docCmd.BoolP("man-page", "", false, "Generate exo manual pages")
	var filesDir = docCmd.StringP("doc-path", "", "./website/content", "Path directory where you want generate doc files")
	var help = docCmd.BoolP("help", "h", false, "Help about any command")

	if err := docCmd.Parse(os.Args); err != nil {
		os.Exit(1)
	}

	if *help {
		_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", os.Args[0], docCmd.FlagUsages())
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	if *manPage {
		header := &doc.GenManHeader{
			Title:   "exo",
			Section: "1",
		}

		err := doc.GenManTree(cmd.RootCmd, header, "./manpage")
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := exoGenMarkdownTreeCustom(cmd.RootCmd, *filesDir); err != nil {
		log.Fatal(err)
	}

}

//
// this following source code is from cobra/doc https://github.com/spf13/cobra/tree/master/doc
// we made that to be able to custom it for our need
//

//
// beginning cobra/doc custom src code
//

func exoGenMarkdownTreeCustom(cmd *cobra.Command, dir string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := os.MkdirAll(filepath.Join(dir, cmd.Name()), 0750); err != nil {
			return err
		}
		if err := exoGenMarkdownTreeCustom(c, filepath.Join(dir, cmd.Name())); err != nil {
			return err
		}
	}

	filename := ""
	if cmd.HasSubCommands() {
		filename = filepath.Join(dir, cmd.Name(), "_index.md")
	} else {
		filename = filepath.Join(dir, cmd.Name()+".md")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck

	fmt.Printf("exo: create file : %s\n", filename)

	return exoGenMarkdownCustom(cmd, f)
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }

// readCommitTime will obtain the commit time out the build info, if any
// If the commit time cannot be found, will return the current time in the same format
func readCommitTime() string {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return time.Now().Format(time.RFC3339)
}

var date string = readCommitTime()

func exoGenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	linkTitle := cmd.Name()
	title := name
	if !cmd.HasParent() {
		linkTitle = "Command Reference"
		title = "Command Reference"
	}
	fmt.Fprintln(buf, "---")
	fmt.Fprintln(buf, "date:", date)
	fmt.Fprintln(buf, "linkTitle:", linkTitle)
	fmt.Fprintln(buf, "title:", title)
	fmt.Fprintln(buf, "description:", cmd.Short)
	fmt.Fprintln(buf, "---")

	long := cmd.Long
	if len(long) != 0 {
		fmt.Fprintln(buf, "### Description")
		fmt.Fprintln(buf)
		fmt.Fprintln(buf, long)
		fmt.Fprintln(buf)
	}

	if cmd.Runnable() {
		fmt.Fprintln(buf, "```")
		fmt.Fprintln(buf, cmd.UseLine())
		fmt.Fprintln(buf, "```")
		fmt.Fprintln(buf)
	} else {
		splitedLine := strings.Split(cmd.UseLine(), " ")

		pos := (len(splitedLine) - 1)

		splitedLine = insertAT(splitedLine, "[command]", pos)

		line := strings.Join(splitedLine, " ")

		fmt.Fprintln(buf, "```")
		fmt.Fprintln(buf, line)
		fmt.Fprintln(buf, "```")
		fmt.Fprintln(buf)
	}

	if len(cmd.Example) > 0 {
		fmt.Fprintln(buf, "### Examples")
		fmt.Fprintln(buf)
		fmt.Fprintln(buf, "```")
		fmt.Fprintln(buf, cmd.Example)
		fmt.Fprintln(buf, "```")
		fmt.Fprintln(buf)
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}
	if hasSeeAlso(cmd) {
		fmt.Fprintln(buf, "### Related Commands")
		fmt.Fprintln(buf)
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.Name()
			link := "../"
			if !cmd.HasSubCommands() {
				link += pname
			}
			fmt.Fprintln(buf, renderRelatedLink(pname, link, parent.Short))
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := child.Name()
			fmt.Fprintln(buf, renderRelatedLink(cname, link, child.Short))
			fmt.Fprintln(buf)
		}
		fmt.Fprintln(buf)
	}

	_, err := buf.WriteTo(w)
	return err
}

func renderRelatedLink(name, link, short string) string {
	return fmt.Sprintf("* [%s]({{< ref \"%s\">}})\t - %s", name, link, short)
}

func tableEscape(s string) string {
	return strings.ReplaceAll(html.EscapeString(s), "|", `\|`)
}

func writeFlag(buf *bytes.Buffer, flag *pflag.Flag) {
	usage := strings.Replace(flag.Usage, "[required]", "(**required**)", 1)
	if flag.Shorthand != "" {
		fmt.Fprintf(buf, "|`--%s, -%s` | %s |\n", flag.Name, flag.Shorthand, tableEscape(usage))
		return
	}
	fmt.Fprintf(buf, "|`--%s` | %s |\n", flag.Name, tableEscape(usage))
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	if flags.HasAvailableFlags() {
		fmt.Fprintln(buf, "### Options")
		fmt.Fprintln(buf)
		fmt.Fprintln(buf, "| Option | Description |")
		fmt.Fprintln(buf, "|---------|------------|")
		flags.VisitAll(func(flag *pflag.Flag) {
			writeFlag(buf, flag)
		})
		fmt.Fprintln(buf)
		fmt.Fprintln(buf)
	}

	parentFlags := cmd.InheritedFlags()
	if parentFlags.HasAvailableFlags() {
		fmt.Fprintln(buf, "### Options inherited from parent commands")
		fmt.Fprintln(buf)
		fmt.Fprintln(buf, "| Option | Description |")
		fmt.Fprintln(buf, "|---------|------------|")
		parentFlags.VisitAll(func(flag *pflag.Flag) {
			writeFlag(buf, flag)
		})
		fmt.Fprintln(buf)
		fmt.Fprintln(buf)
	}
	return nil
}

// Test to see if we have a reason to print See Also information in docs
// Basically this is a test for a parent commend or a subcommand which is
// both not deprecated and not the autogenerated help command.
func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		return true
	}
	return false
}

//
// end cobra/doc custom src code
//

func insertAT(slice []string, elem string, index int) []string {
	return append(slice[:index], append([]string{elem}, slice[index:]...)...)
}
