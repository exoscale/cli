package cmd

import (
	"reflect"
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
)

func TestZoneNames(t *testing.T) {
	got := zoneNames([]v3.ZoneName{"ch-gva-2", "at-vie-1"})
	want := []string{"ch-gva-2", "at-vie-1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("zoneNames() = %v, want %v", got, want)
	}

	if got := zoneNames(nil); got == nil || len(got) != 0 {
		t.Fatalf("zoneNames(nil) = %#v, want an empty slice", got)
	}
}
