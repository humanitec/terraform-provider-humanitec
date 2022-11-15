package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDefinitionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceDefinitionResource("us-east-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "id", "s3-test"),
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "driver_inputs.values.region", "us-east-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_resource_definition.s3_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceDefinitionResource("us-east-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "driver_inputs.values.region", "us-east-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceDefinitionResource(region string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test" {
  id          = "s3-test"
  name        = "s3-test"
  type        = "s3"
  driver_type = "humanitec/s3"

  driver_inputs = {
    values = {
      "region" = "%s"
    }
  }
}
`, region)
}
