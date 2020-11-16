package cmd

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

const (
	templateFilterHelp = "template filter to use (featured|community|mine)"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Templates details",
}

// validateTemplateFilter takes a template filter. Verifies if the
// filter value is valid, and handle filter aliases.
// Returns the template filter to use.
func validateTemplateFilter(templateFilter string) (string, error) {
	if templateFilter == "mine" {
		return "self", nil // nolint
	}
	if templateFilter != "self" && templateFilter != "community" && templateFilter != "featured" {
		return "", fmt.Errorf("invalid template filter %s", templateFilter)
	}
	return templateFilter, nil
}

func getTemplateByNameOrID(zoneID *egoscale.UUID, v string, templateFilter string) (*egoscale.Template, error) {
	req := &egoscale.ListTemplates{
		TemplateFilter: templateFilter,
		ZoneID:         zoneID,
	}

	id, errUUID := egoscale.ParseUUID(v)
	if errUUID != nil {
		req.Name = v
	} else {
		req.ID = id
	}

	resp, err := cs.ListWithContext(gContext, req)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("template %q not found", v)
	}
	if len(resp) == 1 {
		return resp[0].(*egoscale.Template), nil
	}

	// Multiple results returned: we pick the most recent item from the list.
	var (
		template     *egoscale.Template
		templateDate time.Time
	)
	for _, t := range resp {
		ts, err := time.Parse("2006-01-02T15:04:05-0700", t.(*egoscale.Template).Created)
		if err != nil {
			return nil, fmt.Errorf("template creation date parsing error: %s", err)
		}

		if ts.After(templateDate) {
			templateDate = ts
			template = t.(*egoscale.Template)
		}
	}

	return template, nil
}

func findTemplates(zoneID *egoscale.UUID, templateFilter string, filters ...string) ([]egoscale.Template, error) {
	allOS := make(map[string]*egoscale.Template)

	reLinux := regexp.MustCompile(`^Linux (?P<name>.+?) (?P<version>[0-9]+(\.[0-9]+)?)`)
	reVersion := regexp.MustCompile(`(?P<version>[0-9]+(\.[0-9]+)?)`)

	req := &egoscale.ListTemplates{
		TemplateFilter: templateFilter,
		ZoneID:         zoneID,
		Keyword:        strings.Join(filters, " "),
	}

	var err error
	cs.PaginateWithContext(gContext, req, func(i interface{}, e error) bool {
		if e != nil {
			err = e
			return false
		}
		template := i.(*egoscale.Template)
		size := template.Size >> 30 // Size in GiB

		if strings.HasPrefix(template.Name, "Linux") {
			m := reSubMatchMap(reLinux, template.DisplayText)
			if len(m) > 0 {
				if size > 10 {
					// Skipping big, legacy images
					return true
				}

				version, errParse := strconv.ParseFloat(m["version"], 64)
				if errParse != nil {
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
		}

		if strings.HasPrefix(template.Name, "Windows Server") || strings.HasPrefix(template.Name, "OpenBSD") {
			m := reSubMatchMap(reVersion, template.DisplayText)
			if len(m) > 0 {
				version, errParse := strconv.ParseFloat(m["version"], 64)
				if errParse != nil {
					log.Printf("Malformed Windows/OpenBSD version. %q", template.Name)
					return true
				}
				key := fmt.Sprintf("%s %.5f %5d", template.Name[:7], 10000-version, size)
				allOS[key] = template
				return true
			}
		}

		// In doubt, use it directly
		allOS[template.ID.String()] = template
		return true
	})
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(allOS))
	for k := range allOS {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	reDate := regexp.MustCompile(`.* \((?P<date>.*)\)$`)

	templates := make([]egoscale.Template, len(keys))
	for i, k := range keys {
		t := allOS[k]
		m := reSubMatchMap(reDate, t.DisplayText)
		if m["date"] != "" {
			t.Created = m["date"]
		}
		templates[i] = *t
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
	vmCmd.AddCommand(templateCmd)
}
