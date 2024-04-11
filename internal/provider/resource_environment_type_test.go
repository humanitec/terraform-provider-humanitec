package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceEnvironmentType(t *testing.T) {
	id := fmt.Sprintf("qa-env-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceEnvironmentType(id, "Primary QA env"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type.qa", "id", id),
					resource.TestCheckResourceAttr("humanitec_environment_type.qa", "description", "Primary QA env"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_environment_type.qa",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
			// Update testing
			{
				Config: testAccResourceEnvironmentType(id, "Custom QA env"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type.qa", "id", id),
					resource.TestCheckResourceAttr("humanitec_environment_type.qa", "description", "Custom QA env"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceEnvironmentType(id, description string) string {
	return fmt.Sprintf(`
resource "humanitec_environment_type" "qa" {
  id            = "%s"
  description   = "%s"
}
`, id, description)
}
