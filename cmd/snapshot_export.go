package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type snapshotExportOutput struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

func (o *snapshotExportOutput) toJSON()  { output.JSON(o) }
func (o *snapshotExportOutput) toText()  { output.Text(o) }
func (o *snapshotExportOutput) toTable() { output.Table(o) }

var snapshotExportCmd = &cobra.Command{
	Use:   "export ID",
	Short: "Export snapshot",
	Long: fmt.Sprintf(`This command exports a volume snapshot.

Supported output template annotations: %s`,
		strings.Join(output.output.OutputterTemplateAnnotations(&snapshotExportOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		filePath, err := cmd.Flags().GetString("download")
		if err != nil {
			return err
		}

		snapshot, err := exportSnapshot(args[0])
		if err != nil {
			return err
		}

		if !cmd.Flags().Changed("download") {
			return printOutput(&snapshotExportOutput{
				URL:      snapshot.PresignedURL,
				Checksum: snapshot.MD5sum,
			}, nil)
		}

		filePath, err = downloadExportedSnapshot(filePath, snapshot.PresignedURL)
		if err != nil {
			return err
		}

		if !gQuiet {
			fmt.Print("Verifying downloaded file checksum... ")
		}
		if err = checkExportedSnapshot(filePath, snapshot.MD5sum); err != nil {
			if !gQuiet {
				fmt.Println("failed")
			}
			return err
		}

		if !gQuiet {
			fmt.Println("success")
		}

		return nil
	},
}

func exportSnapshot(snapshotID string) (*egoscale.ExportSnapshotResponse, error) {
	id, err := egoscale.ParseUUID(snapshotID)
	if err != nil {
		return nil, err
	}

	res, err := asyncRequest(&egoscale.ExportSnapshot{ID: id}, fmt.Sprintf("Exporting snapshot %q", id))
	if err != nil {
		return nil, err
	}

	return res.(*egoscale.ExportSnapshotResponse), nil
}

func downloadExportedSnapshot(filePath, url string) (string, error) {
	filePath = filepath.ToSlash(filePath)

	st, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if st != nil && st.IsDir() {
		return "", errors.New("download path must not be an existing directory")
	}

	if filepath.Ext(filePath) != ".qcow2" {
		filePath = filePath + ".qcow2"
	}
	if _, err = os.Stat(filePath); err == nil {
		return "", fmt.Errorf("file %q already exists", filePath)
	}

	if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return "", err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(gContext, "GET", url, nil)
	if err != nil {
		return "", err
	}

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return "", err
	}

	progress := mpb.NewWithContext(gContext,
		mpb.WithWidth(64),
		mpb.WithRefreshRate(180*time.Millisecond),
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool { return gQuiet }),
	)

	bar := progress.AddBar(
		int64(size),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name("Downloading snapshot file... "),
			decor.OnComplete(decor.CountersKibiByte("% .2f / % .2f"), "success"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 90), ""),
			decor.OnComplete(decor.Name(""), ""),
		),
	)

	proxyReader := bar.ProxyReader(resp.Body)
	defer proxyReader.Close()

	if _, err = io.Copy(file, proxyReader); err != nil {
		return "", err
	}

	progress.Wait()

	if err = file.Close(); err != nil {
		return "", err
	}

	return filePath, nil
}

func checkExportedSnapshot(filePath, md5sum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	h := hex.EncodeToString(hash.Sum(nil))

	if h != md5sum {
		return fmt.Errorf("checksum mismatch: expected %q, got %q", md5sum, h)
	}

	return nil
}

func init() {
	snapshotCmd.AddCommand(snapshotExportCmd)
	snapshotExportCmd.Flags().StringP("download", "d", "", "Path to download exported snapshot")
}
