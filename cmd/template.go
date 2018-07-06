package cmd

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Templates informations",
}

func getTemplateIDByName(cs *egoscale.Client, name, zoneID string) (string, error) {
	templates, err := cs.List(&egoscale.Template{IsFeatured: true, ZoneID: zoneID})
	if err != nil {
		return "", err
	}

	for _, template := range templates {
		t := template.(*egoscale.Template)
		if name == t.ID {
			return t.ID, nil
		}
	}

	sortedTemplates, err := listTemplates(name)
	if err != nil {
		return "", err
	}

	if len(sortedTemplates) > 1 {
		return "", fmt.Errorf("more than one templates found")
	}
	if len(sortedTemplates) == 1 {
		return sortedTemplates[0].ID, nil
	}

	return "", fmt.Errorf("template %q not found", name)
}

func listTemplates(keywords string) ([]*egoscale.Template, error) {
	zoneID, err := getZoneIDByName(cs, gCurrentAccount.DefaultZone)
	if err != nil {
		return nil, err
	}

	allOS := make(map[string]*egoscale.Template)

	reLinux := regexp.MustCompile(`^Linux (?P<name>.+?) (?P<version>[0-9]+(\.[0-9]+)?)`)
	reVersion := regexp.MustCompile(`(?P<version>[0-9]+(\.[0-9]+)?)`)

	req := &egoscale.ListTemplates{TemplateFilter: "featured", ZoneID: zoneID, Keyword: keywords}

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

	templates := []*egoscale.Template{}
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
		templates = append(templates, &egoscale.Template{
			Name:    t.Name,
			Size:    sz,
			Created: m["date"],
			ID:      t.ID,
		})
	}
	return templates, nil
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
	RootCmd.AddCommand(templateCmd)
}
