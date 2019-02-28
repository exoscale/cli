// +build windows

package utils

import (
	"fmt"
	"os"
)

// PrintJSON prints the given string to stdout.
func PrintJSON(out, theme string) {
	fmt.Fprintln(os.Stdout, out)
}
