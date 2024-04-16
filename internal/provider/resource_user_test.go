package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceUser(t *testing.T) {
	const (
		name     = "test user"
		role     = "member"
		newRole  = "administrator"
		userType = "service"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCreateResourceUser(name, role, userType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_user.test", "name", name),
					resource.TestCheckResourceAttr("humanitec_user.test", "role", role),
					resource.TestCheckResourceAttr("humanitec_user.test", "type", userType),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_user.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := s.RootModule().Resources["humanitec_user.test"].Primary.Attributes["id"]
					return id, nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCreateResourceUser(name, newRole, userType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_user.test", "name", name),
					resource.TestCheckResourceAttr("humanitec_user.test", "role", newRole),
					resource.TestCheckResourceAttr("humanitec_user.test", "type", userType),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCreateResourceUser(name, role, userType string) string {
	return fmt.Sprintf(`
resource "humanitec_user" "test" {
	name = "%s"
	role = "%s"
	type = "%s"
}
`, name, role, userType)
}
