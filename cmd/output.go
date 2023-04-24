package cmd

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

// outputter is an interface that must to be implemented by the commands output
// objects. In addition to the methods, types implementing this interface can
// also use struct tags to modify the output logic:
//   - output:"-" is similar to package encoding/json, i.e. that a field with
//     this tag will not be displayed
//   - outputLabel:"..." overrides the string displayed as label, which by
//     default is the field's CamelCase named split with spaces
type outputter interface {
	toTable()
	toJSON()
	toText()
}

// output prints an outputter interface to the terminal, formatted according
// to the global format specified as CLI flag.
func printOutput(o outputter, err error) error {
	if err != nil {
		return err
	}

	if o == nil {
		return nil
	}

	if output.GOutputTemplate != "" {
		o.toText()
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

// decorateAsyncOperation is a cosmetic helper intended for wrapping long
// asynchronous operations, outputting progress feedback to the user's
// terminal.
func decorateAsyncOperation(message string, fn func()) {
	p := mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(1),
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool { return gQuiet }),
	)

	spinner := p.AddSpinner(
		1,
		mpb.SpinnerOnLeft,
		mpb.AppendDecorators(
			decor.Name(message, decor.WC{W: len(message) + 1, C: decor.DidentRight}),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
		mpb.BarOnComplete("✔"),
	)

	done := make(chan struct{})
	defer close(done)
	go func(doneCh chan struct{}) {
		fn()
		doneCh <- struct{}{}
	}(done)

	<-done
	spinner.Increment(1)
	p.Wait()
}

// proxyWriterAt is a variant of the internal mpb.proxyWriterTo struct,
// required for using mpb with s3manager batch download manager (accepting
// a io.WriterAt interface) since mpb.Bar's ProxyReader() method only
// supports io.Reader and io.WriterTo interfaces.
type proxyWriterAt struct {
	wt  io.WriterAt
	bar *mpb.Bar
	iT  time.Time
}

func (prox *proxyWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = prox.wt.WriteAt(p, off)
	if n > 0 {
		prox.bar.IncrInt64(int64(n), time.Since(prox.iT))
		prox.iT = time.Now()
	}

	if err == io.EOF {
		go prox.bar.SetTotal(0, true)
	}

	return n, err
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

	$ exo config list --output-template '{{ .Name }}' | sort
	alice
	bob

The templating format is Go's text/template, which allows conditional
formatting. For example to display a "*" next to the default configuration
account:

	$ exo config list --output-template '{{ .Name }}{{ if .Default }}*{{ end }}'
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
`,
	},
	)
}

func Int64PtrFormatOutput(n *int64) string {
	if n != nil {
		return strconv.FormatInt(*n, 10)
	}

	return "n/a"
}

func StrPtrFormatOutput(s *string) string {
	if s != nil {
		return *s
	}

	return "n/a"
}

func DatePtrFormatOutput(t *time.Time) string {
	if t != nil {
		return t.String()
	}

	return "n/a"
}
