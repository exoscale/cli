package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
)

const (
	defaultSOSZone = "ch-dk-2"
	minioMaxRetry  = 2
)

// sosCmd represents the sos command
var long = func() string {
	var long = "Manage Exoscale Object Storage (SOS)"

	if runtime.GOOS == "windows" {
		long += `
IMPORTANT: Due to a bug in the Microsoft Windows support in the Go
programming language ( https://github.com/golang/go/issues/16736 ) Windows
users are required to extract the sos-certs.pem file next to their exo.exe
file from the archive. You can obtain a fresh copy of the exo cli from
this address:

    https://github.com/exoscale/cli/releases

The required file can also be obtained from the following address:

	https://www.exoscale.com/static/files/sos-certs.pem

If you have located your certificate chain in a different location you
can also use the '--certs-file' parameter to indicate the location.

We apologize for the inconvenience.
`
	}
	return long
}

var sosCmd = &cobra.Command{
	Use:              "sos",
	Short:            "Simple Object Storage management",
	Long:             long(),
	TraverseChildren: true,
}

type sosClient struct {
	*minio.Client

	certPool *x509.CertPool
}

func newSOSClient(certsFile string) (*sosClient, error) {
	var (
		c   sosClient
		err error
	)

	if certsFile == "" && runtime.GOOS == "windows" {
		// Check if the directory of the "exo" executable contains a file named "sos-certs.pem"
		// to load the certificate chain from. This is done to work around Golang issue #16736
		// on Windows ( https://github.com/golang/go/issues/16736 )
		path, err := os.Executable()
		if err == nil {
			dir, err := filepath.Abs(filepath.Dir(path))
			if err == nil {
				tmpCertsFile := filepath.Join(dir, "sos-certs.pem")
				fmt.Println(tmpCertsFile)
				stat, err := os.Stat(tmpCertsFile)
				if err == nil && stat.IsDir() == false {
					certsFile = tmpCertsFile
				} else {
					log.Warn(
						"It seems you are running on Windows and your sos-certs.pem file is missing.\n" +
							"Please download and extract all files from the exo cli release, not just the executable.\n" +
							"Please see the 'exo sos help' command for more details.")
				}
			}
		}
	}

	if certsFile != "" {
		c.certPool = x509.NewCertPool()
		certs, err := ioutil.ReadFile(certsFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read certificates from file: %s", err)
		}
		if !c.certPool.AppendCertsFromPEM(certs) {
			return nil, errors.New("unable to load local certificates")
		}
	}

	z := gCurrentAccount.DefaultZone
	if z == "" {
		z = defaultSOSZone
	}

	if err = c.setZone(z); err != nil {
		return nil, err
	}

	_, ok := os.LookupEnv("EXOSCALE_TRACE")
	if ok {
		c.TraceOn(os.Stderr)
	}

	return &c, nil
}

func (s *sosClient) setZone(zone string) error {
	// When a user wants to set the SOS zone to use for an operation, we actually have to re-create the
	// underlying Minio S3 client to specify the zone-based endpoint.

	endpoint := strings.Replace(gCurrentAccount.SosEndpoint, "https://", "", -1)
	endpoint = strings.Replace(endpoint, "{zone}", zone, -1)

	minioClient, err := minio.NewV4(endpoint, gCurrentAccount.Key, gCurrentAccount.APISecret(), true)
	if err != nil {
		return err
	}

	// This is a workaround to support SOS on the Windows platform because of a bug preventing access to system-wide
	// trusted CA certificates with Go:
	//   - https://github.com/golang/go/issues/16736
	//   - https://golang.org/src/crypto/x509/root_windows.go#L228
	//
	// Pending resolution, we have to inject the user-provided PEM certificates chain for the TLS SOS API endpoints
	// in our HTTPS client to avoid "https://sos-<zone>.exo.io/: x509: certificate signed by unknown authority" error.
	if s.certPool != nil {
		customTransport, err := func() (http.RoundTripper, error) {
			tr := &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          1024,
				MaxIdleConnsPerHost:   1024,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableCompression:    true,
				TLSClientConfig: &tls.Config{
					RootCAs:    s.certPool,
					MinVersion: tls.VersionTLS12,
				},
			}

			if err := http2.ConfigureTransport(tr); err != nil {
				return nil, err
			}

			return tr, nil
		}()
		if err != nil {
			return fmt.Errorf("unable to initialize custom HTTP transport: %s", err)
		}

		minioClient.SetCustomTransport(customTransport)
	}

	minioClient.SetAppInfo("Exoscale-CLI", gVersion)

	if _, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
		minioClient.TraceOn(os.Stderr)
	}

	s.Client = minioClient

	return nil
}

func init() {
	minio.MaxRetry = minioMaxRetry

	RootCmd.AddCommand(sosCmd)
	sosCmd.PersistentFlags().String("certs-file", "", "Path to file containing additional SOS API X.509 certificates")
}
