package provider

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceWorkloadProfileChartVersion(t *testing.T) {
	chartVersionID := fmt.Sprintf("version-%d", time.Now().UnixNano())
	chartVersionVersion := "1.0.0"

	dir, err := os.MkdirTemp("", "tph-chart-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a minimal helm chart
	chartYAML := fmt.Sprintf("%s/Chart.yaml", dir)
	assert.NoError(t, os.WriteFile(chartYAML, []byte(fmt.Sprintf(`
apiVersion: v2
name: %s
version: %s
`, chartVersionID, chartVersionVersion)), 0644))

	// Create a tar.gz file
	f, err := os.CreateTemp("", "tph-chart-test.tar.gz")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())

	defer os.Remove(f.Name())

	assert.NoError(t, compressDirectory(dir, f.Name()))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceWorkloadProfileChartVersion(f.Name()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_workload_profile_chart_version.main", "id", chartVersionID),
					resource.TestCheckResourceAttr("humanitec_workload_profile_chart_version.main", "version", chartVersionVersion),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_workload_profile_chart_version.main",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", chartVersionID, chartVersionVersion), nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"filename",
					"source_code_hash",
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceWorkloadProfileChartVersion(file string) string {
	return fmt.Sprintf(`
resource "humanitec_workload_profile_chart_version" "main" {
	filename = "%s"
	source_code_hash = filebase64sha256("%s")
}
`, file, file)
}

// From https://gist.github.com/mimoo/25fc9716e0f1353791f5908f94d6e726
// compress a directory into a tar.gz file
func compressDirectory(dir string, archiveName string) error {
	fileToWrite, err := os.OpenFile(archiveName, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
	if err != nil {
		return fmt.Errorf("opening file to write: %w", err)
	}

	if err := compress(dir, fileToWrite); err != nil {
		return fmt.Errorf("compressing directory: %w", err)
	}

	return nil
}

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	base := filepath.Base(src)

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativeFile := strings.Replace(file, src, base, 1)

		// generate tar header
		header, err := tar.FileInfoHeader(fi, relativeFile)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(relativeFile)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}
