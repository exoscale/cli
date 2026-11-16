package storage

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/storage/sos/object"
)

func TestListTimeFilters(t *testing.T) {
	now := time.Now()
	objectAt := func(t time.Time) object.ObjectInterface {
		return &object.Object{Object: &types.Object{Key: aws.String("k"), LastModified: aws.Time(t)}}
	}
	// Objects far from now, so that the duration filters don't depend on when the test runs.
	old := objectAt(now.Add(-48 * time.Hour))
	recent := objectAt(now.Add(-time.Minute))
	ts := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	before := objectAt(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
	after := objectAt(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name               string
		olderThan          time.Duration
		olderThanTimestamp time.Time
		newerThan          time.Duration
		newerThanTimestamp time.Time
		wantFilters        int
		// object -> expected result of all the filters
		want map[object.ObjectInterface]bool
	}{
		{name: "none", wantFilters: 0, want: map[object.ObjectInterface]bool{old: true, recent: true}},
		{
			name: "older than duration", olderThan: 24 * time.Hour, wantFilters: 1,
			want: map[object.ObjectInterface]bool{old: true, recent: false},
		},
		{
			name: "newer than duration", newerThan: 24 * time.Hour, wantFilters: 1,
			want: map[object.ObjectInterface]bool{old: false, recent: true},
		},
		{
			name: "older than timestamp", olderThanTimestamp: ts, wantFilters: 1,
			want: map[object.ObjectInterface]bool{before: true, after: false},
		},
		{
			name: "newer than timestamp", newerThanTimestamp: ts, wantFilters: 1,
			want: map[object.ObjectInterface]bool{before: false, after: true},
		},
		{
			name: "all combined", olderThan: time.Hour, olderThanTimestamp: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			newerThan: 50 * 365 * 24 * time.Hour, newerThanTimestamp: ts, wantFilters: 4,
			want: map[object.ObjectInterface]bool{before: false, after: true, old: true, recent: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := listTimeFilters(
				tt.olderThan, tt.olderThanTimestamp, tt.newerThan, tt.newerThanTimestamp)
			require.Len(t, filters, tt.wantFilters)

			for obj, want := range tt.want {
				require.Equal(t, want, object.ApplyFilters(obj, filters), "object dated %s", obj.GetLastModified())
			}
		})
	}
}

func TestListVersionFilters(t *testing.T) {
	version := func(id string, number uint64) object.ObjectVersionInterface {
		return &object.ObjectVersion{
			ObjectVersion: &types.ObjectVersion{Key: aws.String("k"), VersionId: aws.String(id)},
			VersionNumber: number,
		}
	}
	v0, v1, v2 := version("100", 0), version("101", 1), version("102", 2)

	tests := []struct {
		name        string
		only        []string
		exclude     []string
		wantFilters int
		// version -> expected result of the filters
		want map[object.ObjectVersionInterface]bool
	}{
		{name: "none", want: map[object.ObjectVersionInterface]bool{v0: true, v1: true, v2: true}},
		{
			name: "only ids", only: []string{"100", "102"}, wantFilters: 1,
			want: map[object.ObjectVersionInterface]bool{v0: true, v1: false, v2: true},
		},
		{
			name: "only numbers", only: []string{"v1"}, wantFilters: 1,
			want: map[object.ObjectVersionInterface]bool{v0: false, v1: true, v2: false},
		},
		{
			name: "exclude ids and numbers", exclude: []string{"100", "v2"}, wantFilters: 1,
			want: map[object.ObjectVersionInterface]bool{v0: false, v1: true, v2: false},
		},
		{
			name: "only takes precedence", only: []string{"101"}, exclude: []string{"101"}, wantFilters: 1,
			want: map[object.ObjectVersionInterface]bool{v0: false, v1: true, v2: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := listVersionFilters(tt.only, tt.exclude)
			require.Len(t, filters, tt.wantFilters)

			for v, want := range tt.want {
				require.Equal(t, want, object.ApplyVersionedFilters(v, filters), "version %s", *v.GetVersionID())
			}
		})
	}
}
