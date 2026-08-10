package compute

import "fmt"

const gibibyte int64 = 1 << 30

// ResolveTemplateDiskSize returns a disk size that can contain the selected template.
func ResolveTemplateDiskSize(diskSize, templateSizeBytes int64, explicitlySet bool) (int64, error) {
	minimumDiskSize := (templateSizeBytes + gibibyte - 1) / gibibyte
	if diskSize >= minimumDiskSize {
		return diskSize, nil
	}
	if explicitlySet {
		return 0, fmt.Errorf(
			"--disk-size %d GiB is smaller than the selected template's minimum of %d GiB; use --disk-size %d or larger",
			diskSize,
			minimumDiskSize,
			minimumDiskSize,
		)
	}

	return minimumDiskSize, nil
}
