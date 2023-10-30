package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceApplicationUser(t *testing.T) {
	id := fmt.Sprintf("app-user-test-%d", time.Now().UnixNano())
	testUserID := "1b305f15-f18f-4357-8311-01f88ed99d1b"

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
			resourceName := "humanitec_application_user"
			if tt.deprecatedResource {
				resourceName = "humanitec_resource_application_user"
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: testAccResourceApplicationUser(id, testUserID, "owner", resourceName),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName+".another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
							resource.TestCheckResourceAttr(resourceName+".another_user", "role", "owner"),
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
						Config: testAccResourceApplicationUser(id, testUserID, "developer", resourceName),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName+".another_user", "role", "developer"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceApplicationUser(id, userID, role, resourceName string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "app_user_test" {
	id   = "%s"
	name = "%s"
}

resource "%s" "another_user" {
  app_id  = humanitec_application.app_user_test.id
  user_id = "%s"
  role    = "%s"
}
`, id, id, resourceName, userID, role)
}
