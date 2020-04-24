package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/exoscale/cli/table"
	"github.com/fatih/camelcase"
	"github.com/spf13/cobra"
)

// outputter is an interface that must to be implemented by the commands output
// objects. In addition to the methods, types implementing this interface can
// also use struct tags to modify the output logic:
//   * output:"-" is similar to package encoding/json, i.e. that a field with
//     this tag will not be displayed
//   * outputLabel:"..." overrides the string displayed as label, which by
//     default is the field's CamelCase named split with spaces
type outputter interface {
	toTable()
	toJSON()
	toText()
}

// output prints an outputter interface to the terminal, formatted according
// to the global format specified as CLI flag.
func output(o outputter, err error) error {
	if err != nil {
		return err
	}

	if o == nil {
		return nil
	}

	switch gOutputFormat {
	case "json":
		o.toJSON()

	case "text":
		o.toText()

	default:
		o.toTable()
	}

	return nil
}

// outputterTemplateAnnotations returns a list of annotations available for use
// with an output template.
func outputterTemplateAnnotations(o interface{}) []string {
	annotations := make([]string, 0)

	v := reflect.ValueOf(o)
	v = reflect.Indirect(v)
	t := v.Type()

	// If the outputter interface is iterable (slice only), use the element type
	if v.Kind() == reflect.Slice {
		t = v.Type().Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		annotations = append(annotations, "."+t.Field(i).Name)
	}

	return annotations
}

// outputJSON prints a JSON-formatted rendering of o to the terminal.
func outputJSON(o interface{}) {
	j, err := json.Marshal(o)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to encode output to JSON: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(j))
}

// outputText prints a template-based plain text rendering of o to the
// terminal. If the object is of iterable type (slice only), each item is
// printed on a new line. If none is provided by the user, the default
// template prints all fields separated by a tabulation character.
func outputText(o interface{}) {
	var tpl = gOutputTemplate

	if tpl == "" {
		tplFields := outputterTemplateAnnotations(o)
		for i := range tplFields {
			tplFields[i] = "{{" + tplFields[i] + "}}"
		}
		tpl = strings.Join(tplFields, "\t")
	}

	t, err := template.New("out").Parse(tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to encode output in plaintext using template: %s\n", err)
		os.Exit(1)
	}

	// If the outputter interface is iterable (slice only), we loop over the
	// items and perform the templating directly
	if v := reflect.ValueOf(o); reflect.Indirect(v).Kind() == reflect.Slice {
		for i := 0; i < reflect.Indirect(v).Len(); i++ {
			if err := t.Execute(os.Stdout, reflect.Indirect(v).Index(i).Interface()); err != nil {
				fmt.Fprintf(os.Stderr, "error: unable to encode output using template: %s\n", err)
				os.Exit(1)
			}
			fmt.Println()
		}
		return
	}

	if err := t.Execute(os.Stdout, o); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to encode output using template: %s\n", err)
		os.Exit(1)
	}

	fmt.Println()
}

// outputTable prints a table-formatted rendering of o to the terminal.
// If the object is of iterable type (slice only), each item is printed in a
// table row, with a header containing one column per type field. Otherwise,
// each field of the object is printed in a key/value formatted table, and a
// header is printed if the item type implements an optional (Type() string)
// method.
func outputTable(o interface{}) {
	tab := table.NewTable(os.Stdout)

	v := reflect.ValueOf(o)
	v = reflect.Indirect(v)
	t := v.Type()

	// If the outputter interface is iterable (slice only), use the element type
	if v.Kind() == reflect.Slice {
		t = v.Type().Elem()
	}

	// Turn CamelCase field names into eye-friendlier labels.
	// If the field has an `outputLabel` tag, use its value to override the header label.
	headers := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		// Check if the field has to be skipped
		if l, ok := t.Field(i).Tag.Lookup("output"); ok {
			if l == "-" {
				continue
			}
		}

		label := strings.Join(camelcase.Split(t.Field(i).Name), " ")
		if l, ok := t.Field(i).Tag.Lookup("outputLabel"); ok {
			label = l
		}
		headers = append(headers, label)
	}

	// If the outputter interface is iterable (slice only), we loop over the
	// items and display each one in a table row
	if v := reflect.ValueOf(o); reflect.Indirect(v).Kind() == reflect.Slice {
		tab.SetHeader(headers)

		for i := 0; i < reflect.Indirect(v).Len(); i++ {
			item := reflect.Indirect(v).Index(i)
			row := make([]string, 0)

			for j := 0; j < item.NumField(); j++ {
				field := item.Field(j)
				// Check if the field has to be skipped
				if l, ok := item.Type().Field(j).Tag.Lookup("output"); ok {
					if l == "-" {
						continue
					}
				}

				switch field.Kind() {
				case reflect.Slice:
					// If the field value is a slice and is empty,
					// print "n/a" instead of an empty slice
					if field.Len() == 0 {
						row = append(row, "n/a")
					} else {
						row = append(row, fmt.Sprint(field.Interface()))
					}

				case reflect.Ptr:
					// If the field value is a nil pointer, print "n/a" instead of <nil>
					if field.IsNil() {
						row = append(row, "n/a")
					} else {
						row = append(row, fmt.Sprint(field.Interface()))
					}

				default:
					row = append(row, fmt.Sprint(field.Interface()))
				}
			}

			tab.Append(row)
		}

		tab.Render()
		return
	}

	// Single item, loop over the type fields and display each item in a key/value-type table

	// If the outputter interface implements the optional `Type()` method,
	// use its return value as table header
	if typeMethod := reflect.ValueOf(o).MethodByName("Type"); typeMethod.Kind() != reflect.Invalid {
		in := make([]reflect.Value, typeMethod.Type().NumIn())
		header := typeMethod.Call(in)[0].Interface().(string)
		tab.SetHeader([]string{header, ""})
	}

	for i := 0; i < t.NumField(); i++ {
		// Check if the field has to be skipped
		if l, ok := t.Field(i).Tag.Lookup("output"); ok {
			if l == "-" {
				continue
			}
		}

		label := strings.Join(camelcase.Split(t.Field(i).Name), " ")
		if l, ok := t.Field(i).Tag.Lookup("outputLabel"); ok {
			label = l
		}

		switch v.Field(i).Kind() {
		case reflect.Slice:
			// If the field value is a slice and is empty, print "n/a" instead of 0
			if n := v.Field(i).Len(); n == 0 {
				tab.Append([]string{label, "n/a"})
			} else {
				items := v.Field(i).Interface().([]string)
				tab.Append([]string{label, strings.Join(items, "\n")})
			}

		case reflect.Ptr:
			// If the field value is a nil pointer, print "n/a" instead of <nil>
			if v.Field(i).IsNil() {
				tab.Append([]string{label, "n/a"})
			} else {
				tab.Append([]string{label, fmt.Sprint(v.Field(i).Interface())})
			}

		default:
			tab.Append([]string{label, fmt.Sprint(v.Field(i).Interface())})
		}
	}

	tab.Render()
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

	$ exo config list -O text --output-template '{{ .Name }}' | sort
	alice
	bob

The templating format is Go's text/template, which allows conditional
formatting. For example to display a "*" next to the default configuration
account:

	$ exo config list -O text \
		--output-template '{{ .Name }}{{ if .Default }}*{{ end }}'
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
`},
	)
}
