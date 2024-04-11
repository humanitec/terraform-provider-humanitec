package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceEnvironmentTypeUser(t *testing.T) {
	id := fmt.Sprintf("env-type-user-test-%d", time.Now().UnixNano())
	testUserID := "c0725726-0613-43d4-8398-907d07fba2e4"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "role", "deployer"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_environment_type_user.another_user",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", id, testUserID), nil
				},
			},
			// Update and Read testing
			{
				// At the moment, there is nothing we can update :-/
				Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "role", "deployer"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceEnvironmentTypeUser(id, userID, role string) string {
	return fmt.Sprintf(`
resource "humanitec_environment_type" "qa" {
	id            = "%s"
	description   = "%s"
}

resource "humanitec_environment_type_user" "another_user" {
  env_type_id = humanitec_environment_type.qa.id
  user_id     = "%s"
  role        = "%s"
}
`, id, id, userID, role)
}
