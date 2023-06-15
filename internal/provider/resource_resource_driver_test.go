package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceResourceDriver(t *testing.T) {
	tests := []struct {
		name         string
		configCreate func(id string) string
		configUpdate func(id string) string
		testCreate   resource.TestCheckFunc
		testUpdate   resource.TestCheckFunc
	}{
		{
			name: "basic",
			configCreate: func(id string) string {
				return testAccResourceResourceDriver(id, "https://drivers.example.com/s3/")
			},
			configUpdate: func(id string) string {
				return testAccResourceResourceDriver(id, "https://drivers.example.com/s3-new/")
			},
			testCreate: resource.TestCheckResourceAttr("humanitec_resource_driver.s3", "target", "https://drivers.example.com/s3/"),
			testUpdate: resource.TestCheckResourceAttr("humanitec_resource_driver.s3", "target", "https://drivers.example.com/s3-new/"),
		},
		{
			name: "virtual",
			configCreate: func(id string) string {
				return testAccResourceResourceDriverVirtual(id, "\"static\"")
			},
			configUpdate: func(id string) string {
				return testAccResourceResourceDriverVirtual(id, "{ \"type\" = \"static\" }")
			},
			testCreate: resource.TestCheckResourceAttr("humanitec_resource_driver.s3", "template", "\"static\""),
			testUpdate: resource.TestCheckResourceAttr("humanitec_resource_driver.s3", "template", "{\"type\":\"static\"}"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprintf("driver-%d", time.Now().UnixNano())

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: tc.configCreate(id),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_resource_driver.s3", "id", id),
							tc.testCreate,
						),
					},
					// ImportState testing
					{
						ResourceName:      "humanitec_resource_driver.s3",
						ImportState:       true,
						ImportStateVerify: true,
					},
					// Update and Read testing
					{
						Config: tc.configUpdate(id),
						Check: resource.ComposeAggregateTestCheckFunc(
							tc.testUpdate,
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceResourceDriver(id, target string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_driver" "s3" {
	id   = "%s"
	type = "s3"

	account_types = [
		"aws",
	]

	inputs_schema = jsonencode({})
	target        = "%s"
}
`, id, target)
}

func testAccResourceResourceDriverVirtual(id, target string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_driver" "s3" {
	id   = "%s"
	type = "s3"

	account_types = [
		"aws",
	]

	inputs_schema = jsonencode({})
	target        = "driver://humanitec/static"
	template = jsonencode(%s)
}
`, id, target)
}
