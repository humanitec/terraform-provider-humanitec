package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAccountResource(t *testing.T) {
	id := fmt.Sprintf("gcp-test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceAccountResource(id, "gcp-test-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_account.gcp_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_resource_account.gcp_test", "name", "gcp-test-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_resource_account.gcp_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
			// Update and Read testing
			{
				Config: testAccResourceAccountResource(id, "gcp-test-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_account.gcp_test", "name", "gcp-test-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceAccountResource(id, name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_account" "gcp_test" {
  id          = "%s"
  name        = "%s"
  type        = "gcp"
  credentials = "{}"
}
`, id, name)
}
