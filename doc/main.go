package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

const frontmatter = `---
date: %s
title: %q
slug: %q
url: %s
description: %q
type: %s
---
`

func main() {

	var flagError pflag.ErrorHandling
	docCmd := pflag.NewFlagSet("", flagError)
	var isHugo = docCmd.BoolP("is-hugo", "", true, "set false if you dont want to generate fot hugo (https://gohugo.io/)")
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

	filePrepender := func(filename string, cmd *cobra.Command) string {
		now := time.Now().Format(time.RFC3339)
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		url := fmt.Sprintf("/%s/", strings.ToLower(base))
		slug := strings.Replace(base, "_", " ", -1)
		typeExo := `"command"`
		if strings.Count(base, "_") > 1 {
			typeExo = `"subcommand"`
		}
		return fmt.Sprintf(frontmatter, now, slug, base, url, cmd.Short, typeExo)
	}

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return fmt.Sprintf("/cli/%s/", strings.ToLower(base))
	}

	if err := exoGenMarkdownTreeCustom(cmd.RootCmd, *filesDir, filePrepender, linkHandler, *isHugo); err != nil {
		log.Fatal(err)
	}

}

//
// this following source code is from cobra/doc https://github.com/spf13/cobra/tree/master/doc
// we made that to be able to custom it for our need
//

//
//beginning cobra/doc custom src code
//

func exoGenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender func(string, *cobra.Command) string, linkHandler func(string) string, isHugo bool) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := exoGenMarkdownTreeCustom(c, dir, filePrepender, linkHandler, isHugo); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck

	fmt.Printf("exo: create file : %s\n", filename)

	if _, err := io.WriteString(f, filePrepender(filename, cmd)); err != nil {
		return err
	}

	return exoGenMarkdownCustom(cmd, f, linkHandler, isHugo)
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }

func exoGenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string, ishugo bool) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	if !ishugo {
		short := cmd.Short
		buf.WriteString("## " + name + "\n\n")
		buf.WriteString(short + "\n\n")
	}

	if ishugo {
		buf.WriteString("<!--more-->\n\n")
	}

	long := cmd.Long
	if len(long) != 0 {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	} else {
		splitedLine := strings.Split(cmd.UseLine(), " ")

		pos := (len(splitedLine) - 1)

		splitedLine = insertAT(splitedLine, "[command]", pos)

		line := strings.Join(splitedLine, " ")

		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", line))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}
	if hasSeeAlso(cmd) {
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := pname + ".md"
			link = strings.Replace(link, " ", "_", -1)
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, linkHandler(link), parent.Short))
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := cname + ".md"
			link = strings.Replace(link, " ", "_", -1)
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n\n", cname, linkHandler(link), child.Short))
		}
		buf.WriteString("\n")
	}

	_, err := buf.WriteTo(w)
	return err
}

func writeFlag(buffer *bytes.Buffer, flag *pflag.Flag) {
	usage := strings.Replace(flag.Usage, "[required]", "(**required**)", 1)
	if flag.Shorthand != "" {
		buffer.WriteString(fmt.Sprintf("`--%s, -%s` - %s\n", flag.Name, flag.Shorthand, html.EscapeString(usage)))
		return
	}
	buffer.WriteString(fmt.Sprintf("`--%s` - %s\n", flag.Name, html.EscapeString(usage)))
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n")
		flags.VisitAll(func(flag *pflag.Flag) {
			writeFlag(buf, flag)
		})
		buf.WriteString("\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n")
		parentFlags.VisitAll(func(flag *pflag.Flag) {
			writeFlag(buf, flag)
		})
		buf.WriteString("\n\n")
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
//end cobra/doc custom src code
//

func insertAT(slice []string, elem string, index int) []string {
	return append(slice[:index], append([]string{elem}, slice[index:]...)...)
}
