package key

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type successResponseOutput v3.SuccessResponse

func (o *successResponseOutput) ToJSON() { output.JSON(o) }
func (o *successResponseOutput) ToText() { output.Text(o) }
func (o *successResponseOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"STATUS",
	})

	t.Append([]string{
		string(o.Status),
	})
}

func formatKeyRotationConfig(s *v3.KeyRotationConfig) string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("auto: %s\ncount: %d\nnextAt: %s\nrotationPeriod: %d",
		strconv.FormatBool(*s.Automatic),
		s.ManualCount,
		s.NextAT,
		s.RotationPeriod)
}

func formatKeyMaterial(s *v3.KeyMaterial) string {
	if s == nil {
		return "-"
	}
	// TODO: temp fix to prevent null pointer exception. Automatic field will be renamed to manual.
	if s.Automatic != nil {
		return fmt.Sprintf("auto: %s\ncreatedAt: %s\nversion: %d",
			strconv.FormatBool(*s.Automatic),
			s.CreatedAT,
			s.Version)
	}
	return fmt.Sprintf("createdAt: %s\nversion: %d",
		s.CreatedAT,
		s.Version)
}

func formatReplicaStatus(s []v3.ReplicaState) string {
	if len(s) == 0 {
		return "-"
	}
	var res []string
	for _, r := range s {
		res = append(res, r.Zone)
	}
	return strings.Join(res, ", ")
}
