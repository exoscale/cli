package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
)

// output prints an outputter interface to the terminal, formatted according
// to the global format specified as CLI flag.
func PrintOutput(o output.Outputter, err error) error {
	if err != nil {
		return err
	}

	if o == nil {
		return nil
	}

	if output.GOutputTemplate != "" {
		o.ToText()
		return nil
	}

	switch globalstate.OutputFormat {
	case "json":
		o.ToJSON()

	case "text":
		o.ToText()

	default:
		o.ToTable()
	}

	return nil
}

// utils.DecorateAsyncOperation is a cosmetic helper intended for wrapping long
// asynchronous operations, outputting progress feedback to the user's
// terminal.
// TODO remove this one once all has been migrated to utils.DecorateAsyncOperationrationrationrationrationrations.
func DecorateAsyncOperation(message string, fn func()) {
	p := mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(1),
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool { return globalstate.Quiet }),
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

func DecorateAsyncOperations(message string, fns ...func() error) error {
	if len(fns) == 0 {
		return nil
	}

	p := mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(1),
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool { return globalstate.Quiet }),
	)

	spinner := p.AddSpinner(
		int64(len(fns)),
		mpb.SpinnerOnLeft,
		mpb.AppendDecorators(
			decor.Name(message, decor.WC{W: len(message) + 1, C: decor.DidentRight}),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
		mpb.BarOnComplete("✔"),
	)

	errs := &multierror.Error{}
	done := make(chan struct{})
	defer close(done)

	for i := 0; i < len(fns); i += 10 {
		batchSize := min(10, len(fns)-i)
		for j := 0; j < batchSize; j++ {
			fnIndex := i + j
			go func(doneCh chan struct{}, fn func() error) {
				if err := fn(); err != nil {
					errs = multierror.Append(errs, err)
				}
				doneCh <- struct{}{}
			}(done, fns[fnIndex])
		}

		for j := 0; j < batchSize; j++ {
			<-done
			spinner.Increment(1)
		}
	}

	p.Wait()

	return errs.ErrorOrNil()
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
