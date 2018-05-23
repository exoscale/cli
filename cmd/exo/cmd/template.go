package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "list all available templates",
	Run: func(cmd *cobra.Command, args []string) {

		infos, err := listTemplates()
		if err != nil {
			log.Fatal(err)
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Operating System", "Disk", "Release Date", "ID"})

		for _, v := range infos {
			sz := strconv.FormatInt(v.Size, 10)
			if sz == "0" {
				sz = ""
			}
			table.Append([]string{v.Name, sz, v.Created, v.ID})
		}
		table.Render()
	},
}

func getTemplateIDByName(cs *egoscale.Client, name string) (string, error) {
	templates, err := cs.List(&egoscale.Template{IsFeatured: true})
	if err != nil {
		return "", err
	}

	for _, template := range templates {
		t := template.(*egoscale.Template)
		if strings.Compare(strings.ToLower(name), strings.ToLower(t.Name)) == 0 {
			return t.ID, nil
		}
	}
	return name, nil
}

func listTemplates() ([]*egoscale.Template, error) {
	template := &egoscale.Template{IsFeatured: true, ZoneID: "1"}
	req, err := template.ListRequest()
	if err != nil {
		return nil, err
	}

	allOS := make(map[string]*egoscale.Template)

	reLinux := regexp.MustCompile(`^Linux (?P<name>.+?) (?P<version>[0-9]+(\.[0-9]+)?)`)
	reVersion := regexp.MustCompile(`(?P<version>[0-9]+(\.[0-9]+)?)`)

	cs.Paginate(req, func(i interface{}, err error) bool {
		template := i.(*egoscale.Template)
		template.Size = template.Size >> 30 //Size in Gib
		if strings.HasPrefix(template.Name, "Linux") {
			m := reSubMatchMap(reLinux, template.DisplayText)
			if len(m) > 0 {
				if template.Size > 10 {
					// Skipping big, legacy images
					return true
				}

				version, err := strconv.ParseFloat(m["version"], 64)
				if err != nil {
					log.Printf("Malformed Linux version. got %q in %q", m["version"], template.Name)
					return true
				}
				res := fmt.Sprintf("%.5f", 10000-version)

				// fix Container Linux sorting
				name := strings.Replace(m["name"], "stable ", "", 1)
				key := fmt.Sprintf("Linux %s %s", name, res)
				allOS[key] = template
				return true
			}
			// skip
			log.Printf("Malformed Linux. %q", template.DisplayText)
			return true
		}

		if strings.HasPrefix(template.Name, "Windows Server") || strings.HasPrefix(template.Name, "OpenBSD") {
			m := reSubMatchMap(reVersion, template.DisplayText)
			if len(m) > 0 {
				version, err := strconv.ParseFloat(m["version"], 64)
				if err != nil {
					log.Printf("Malformed Windows/OpenBSD version. %q", template.Name)
					return true
				}
				key := fmt.Sprintf("%s %.5f %5d", template.Name[:7], 10000-version, template.Size)
				allOS[key] = template
				return true
			}

			log.Printf("Malformed Windows/OpenBSD. %q", template.DisplayText)
			return true
		}

		// In doubt, use it directly
		allOS[template.Name] = template
		return true
	})

	var keys []string
	for k := range allOS {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	reDate := regexp.MustCompile(`.* \((?P<date>.*)\)$`)

	infos := []*egoscale.Template{}
	for _, k := range keys {
		t := allOS[k]
		m := reSubMatchMap(reDate, t.DisplayText)
		size := fmt.Sprintf("%d", t.Size)
		if strings.HasPrefix(t.DisplayText, "Linux") {
			size = "0"
		}

		sz, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return nil, err
		}
		infos = append(infos, &egoscale.Template{Name: t.Name, Size: sz, Created: m["date"], ID: t.ID})
	}
	return infos, nil
}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 && len(match) > 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
