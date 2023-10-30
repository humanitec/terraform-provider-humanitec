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

	tests := []struct {
		name               string
		deprecatedResource bool
	}{
		{
			name:               "regular",
			deprecatedResource: false,
		},
		{
			name:               "deprecated",
			deprecatedResource: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceName := "humanitec_environment_type_user"
			if tt.deprecatedResource {
				resourceName = "humanitec_resource_environment_type_user"
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer", resourceName),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName+".another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
							resource.TestCheckResourceAttr(resourceName+".another_user", "role", "deployer"),
						),
					},
					// ImportState testing
					{
						ResourceName:      resourceName + ".another_user",
						ImportState:       true,
						ImportStateVerify: true,
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							return fmt.Sprintf("%s/%s", id, testUserID), nil
						},
					},
					// Update and Read testing
					{
						// At the moment there is nothing we can update :-/
						Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer", resourceName),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName+".another_user", "role", "deployer"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceEnvironmentTypeUser(id, userID, role, resourceName string) string {
	return fmt.Sprintf(`
resource "humanitec_environment_type" "qa" {
	id            = "%s"
	description   = "%s"
}

resource "%s" "another_user" {
  env_type_id = humanitec_environment_type.qa.id
  user_id     = "%s"
  role        = "%s"
}
`, id, id, resourceName, userID, role)
}
