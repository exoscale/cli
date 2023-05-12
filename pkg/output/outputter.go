package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"strings"

	"github.com/exoscale/cli/table"

	"github.com/fatih/camelcase"
)

var (
	GOutputTemplate string
)

// JSON prints a JSON-formatted rendering of o to the terminal.
func JSON(o interface{}) {
	j, err := json.Marshal(o)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to encode output to JSON: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(j))
}

// Outputter is an interface that must to be implemented by the commands output
// objects. In addition to the methods, types implementing this interface can
// also use struct tags to modify the output logic:
//   - output:"-" is similar to package encoding/json, i.e. that a field with
//     this tag will not be displayed
//   - outputLabel:"..." overrides the string displayed as label, which by
//     default is the field's CamelCase named split with spaces
type Outputter interface {
	ToTable()
	ToJSON()
	ToText()
}

// Text prints a template-based plain text rendering of o to the
// terminal. If the object is of iterable type (slice only), each item is
// printed on a new line. If none is provided by the user, the default
// template prints all fields separated by a tabulation character.
func Text(o interface{}) {
	tpl := GOutputTemplate

	if tpl == "" {
		tplFields := TemplateAnnotations(o)
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
}

// outputTableHeaders turns CamelCase field names into eye-friendlier labels.
// If the field has an `outputLabel` tag, use its value to override the header label.
func outputTableHeaders(t reflect.Type) []string {
	headers := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		// Check if the field has to be skipped.
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

	return headers
}

// outputTableRow turns the fields of an item into a table row
func outputTableRow(item reflect.Value) []string {
	row := []string{}
	for i := 0; i < item.NumField(); i++ {
		field := item.Field(i)
		// Check if the field has to be skipped.
		if l, ok := item.Type().Field(i).Tag.Lookup("output"); ok {
			if l == "-" {
				continue
			}
		}

		switch field.Kind() {
		case reflect.Slice:
			// If the field value is a slice and is empty,
			// print "n/a" instead of an empty slice.
			if field.Len() == 0 {
				row = append(row, "n/a")
			} else {
				row = append(row, fmt.Sprint(field.Interface()))
			}

		case reflect.Map:
			// If the field value is a map and is empty,
			// print "n/a" instead of an empty map.
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
				row = append(row, fmt.Sprint(field.Elem().Interface()))
			}

		default:
			row = append(row, fmt.Sprint(field.Interface()))
		}
	}

	return row
}

// Table prints a table-formatted rendering of o to the terminal.
// If the object is of iterable type (slice only), each item is printed in a
// table row, with a header containing one column per type field. Otherwise,
// each field of the object is printed in a key/value formatted table, and a
// header is printed if the item type implements an optional (Type() string)
// method.
func Table(o interface{}) {
	tab := table.NewTable(os.Stdout)

	v := reflect.ValueOf(o)
	v = reflect.Indirect(v)
	t := v.Type()

	// If the outputter interface is iterable (slice only), use the element type.
	if v.Kind() == reflect.Slice {
		t = v.Type().Elem()
	}

	headers := outputTableHeaders(t)

	// If the outputter interface is iterable (slice only), we loop over the
	// items and display each one in a table row.
	if v := reflect.ValueOf(o); reflect.Indirect(v).Kind() == reflect.Slice {
		tab.SetHeader(headers)

		for i := 0; i < reflect.Indirect(v).Len(); i++ {
			row := outputTableRow(reflect.Indirect(v).Index(i))
			tab.Append(row)
		}

		tab.Render()
		return
	}

	// Single item, loop over the type fields and display each item in a key/value-type table.

	// If the outputter interface implements the optional `Type()` method,
	// use its return value as table header.
	if typeMethod := reflect.ValueOf(o).MethodByName("Type"); typeMethod.Kind() != reflect.Invalid {
		in := make([]reflect.Value, typeMethod.Type().NumIn())
		header := typeMethod.Call(in)[0].Interface().(string)
		tab.SetHeader([]string{header, ""})
	}

	for i := 0; i < t.NumField(); i++ {
		// Check if the field has to be skipped.
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
			// If the field value is a slice and is empty, print "n/a" instead of 0.
			if n := v.Field(i).Len(); n == 0 {
				tab.Append([]string{label, "n/a"})
			} else {
				switch v.Field(i).Type().Elem().Kind() {
				case reflect.Struct:
					var embeddedBuf bytes.Buffer
					embeddedTable := table.NewEmbeddedTable(&embeddedBuf)

					embeddedTable.SetHeader(outputTableHeaders(v.Field(i).Type().Elem()))

					for j := 0; j < reflect.Indirect(v.Field(i)).Len(); j++ {
						row := outputTableRow(reflect.Indirect(v.Field(i)).Index(j))
						embeddedTable.Append(row)
					}

					embeddedTable.Render()
					tab.Append([]string{label, embeddedBuf.String()})
				case reflect.String:
					items := v.Field(i).Interface().([]string)
					tab.Append([]string{label, strings.Join(items, "\n")})
				default:
					tab.Append([]string{label, "(type not supported)\n"})
				}
			}

		case reflect.Map:
			// If the field value is a map and is empty, print "n/a" instead of 0.
			if n := v.Field(i).Len(); n == 0 {
				tab.Append([]string{label, "n/a"})
			} else {
				items := v.Field(i).Interface().(map[string]string)
				tab.Append([]string{label, strings.Join(func() []string {
					list := make([]string, 0)
					for k, v := range items {
						list = append(list, fmt.Sprintf("%s:%s", k, v))
					}
					return list
				}(), "\n")})
			}

		case reflect.Ptr:
			// If the field value is a nil pointer, print "n/a" instead of <nil>.
			if v.Field(i).IsNil() {
				tab.Append([]string{label, "n/a"})
			} else {
				tab.Append([]string{label, fmt.Sprint(v.Field(i).Elem().Interface())})
			}

		default:
			tab.Append([]string{label, fmt.Sprint(v.Field(i).Interface())})
		}
	}

	tab.Render()
}

// TemplateAnnotations returns a list of annotations available for use
// with an output template.
func TemplateAnnotations(o interface{}) []string {
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
