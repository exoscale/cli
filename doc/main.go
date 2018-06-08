package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/exoscale/egoscale/cmd/exo/cmd"
	"github.com/spf13/cobra/doc"
)

const frontmatter = `---
date: %s
title: %s
slug: %s
url: %s
---
`

func main() {
	filePrepender := func(filename string) string {
		now := time.Now().Format(time.RFC3339)
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		url := fmt.Sprintf("/cli/%s/", strings.ToLower(base))
		return fmt.Sprintf(frontmatter, now, strings.Replace(base, "_", " ", -1), base, url)
	}

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return fmt.Sprintf("/egoscale/cli/%s/", strings.ToLower(base))
	}

	doc.GenMarkdownTreeCustom(cmd.RootCmd, "../../website/content/cli", filePrepender, linkHandler)

}
