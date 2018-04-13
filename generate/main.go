package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/exoscale/egoscale"
)

var cmd = flag.String("cmd", "", "CloudStack command name")
var source = flag.String("apis", "", "listApis response in JSON")

// fieldInfo represents the inner details of a field
type fieldInfo struct {
	Var       *types.Var
	OmitEmpty bool
	Doc       string
}

// command represents a struct within the source code
type command struct {
	name     string
	s        *types.Struct
	position token.Pos
	fields   map[string]fieldInfo
	errors   map[string]error
}

func main() {
	flag.Parse()

	sourceFile, _ := os.Open(*source)
	decoder := json.NewDecoder(sourceFile)
	apis := new(egoscale.ListAPIsResponse)
	if err := decoder.Decode(&apis); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	fset := token.NewFileSet()
	astFiles := make([]*ast.File, 0)
	files, err := filepath.Glob("*.go")
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		astFiles = append(astFiles, f)
	}

	info := types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{
		Importer: importer.For("source", nil),
	}

	_, err = conf.Check("egoscale", fset, astFiles, &info)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	commands := make(map[string]*command)

	for id, obj := range info.Defs {
		if obj == nil || !obj.Exported() {
			continue
		}

		typ := obj.Type().Underlying()

		switch typ.(type) {
		case *types.Struct:
			commands[strings.ToLower(obj.Name())] = &command{
				name:     obj.Name(),
				s:        typ.(*types.Struct),
				position: id.Pos(),
			}
		}
	}

	re := regexp.MustCompile(`\bjson:"(?P<name>[^,"]+)(?P<omit>,omitempty)?"`)
	reDoc := regexp.MustCompile(`\bdoc:"(?P<doc>[^"]+)"`)

	for _, a := range apis.API {
		if cmd, ok := commands[strings.ToLower(a.Name)]; !ok {
			// too much information
			//fmt.Fprintf(os.Stderr, "Unknown command: %q\n", a.Name)
		} else {
			// mapping from name to field
			cmd.fields = make(map[string]fieldInfo)
			cmd.errors = make(map[string]error)

			for i := 0; i < cmd.s.NumFields(); i++ {
				f := cmd.s.Field(i)
				if !f.IsField() || !f.Exported() {
					continue
				}

				tag := cmd.s.Tag(i)
				match := re.FindStringSubmatch(tag)
				if len(match) == 0 {
					cmd.errors[f.Name()] = fmt.Errorf("Field error: no json annotation found")
					continue
				}
				name := match[1]
				omitempty := len(match) == 3 && match[2] == ",omitempty"

				doc := ""
				match = reDoc.FindStringSubmatch(tag)
				if len(match) == 2 {
					doc = match[1]
				}

				cmd.fields[name] = fieldInfo{
					Var:       f,
					OmitEmpty: omitempty,
					Doc:       doc,
				}
			}

			for _, p := range a.Params {
				field, ok := cmd.fields[p.Name]

				if !ok {
					cmd.errors[p.Name] = fmt.Errorf("Field missing")
					continue
				}
				delete(cmd.fields, p.Name)

				typename := field.Var.Type().String()
				expected := ""
				switch p.Type {
				case "integer":
					if typename != "int" {
						expected = "int"
					}
				case "long":
					if typename != "int64" {
						expected = "int64"
					}
				case "boolean":
					if typename != "bool" && typename != "*bool" {
						expected = "bool"
					}
				case "string":
				case "uuid":
					if typename != "string" {
						expected = "string"
					}
				case "map":
					if !strings.HasPrefix(typename, "[]") {
						expected = "array"
					}
				default:
					cmd.errors[p.Name] = fmt.Errorf("Unknown type %q <=> %q", p.Type, field.Var.Type().String())
				}

				if expected != "" {
					cmd.errors[p.Name] = fmt.Errorf("Expected to be a slice[], got %q", typename)
				}

				if field.Doc != p.Description {
					cmd.errors["tag:"+p.Name] = fmt.Errorf("missing `doc:%q`", p.Description)
				}
			}

			for name := range cmd.fields {
				cmd.errors[name] = fmt.Errorf("Extra field found")
			}
		}
	}

	for name, c := range commands {
		pos := fset.Position(c.position)
		er := len(c.errors)

		if *cmd == "" {
			if er != 0 {
				fmt.Printf("%5d %s: %s\n", er, pos, c.name)
			}
		} else if strings.ToLower(*cmd) == name {
			for k, e := range c.errors {
				fmt.Printf("%s: %s\n", k, e.Error())
			}
			fmt.Printf("\n%s: %s has %d error(s)\n", pos, c.name, er)
			os.Exit(er)
		}
	}
}
