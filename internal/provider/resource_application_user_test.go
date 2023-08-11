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
	eventualConsistentUserAPITimeout := 30 * time.Second

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceApplicationUser(id, testUserID, "owner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_application_user.another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
					resource.TestCheckResourceAttr("humanitec_resource_application_user.another_user", "role", "owner"),
				),
			},
			// ImportState testing
			{
				PreConfig: func() {
					time.Sleep(eventualConsistentUserAPITimeout)
				},
				ResourceName:      "humanitec_resource_application_user.another_user",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", id, testUserID), nil
				},
			},
			// Update and Read testing
			{
				Config: testAccResourceApplicationUser(id, testUserID, "developer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(eventualConsistentUserAPITimeout)
						return nil
					},
					resource.TestCheckResourceAttr("humanitec_resource_application_user.another_user", "role", "developer"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceApplicationUser(id, userID, role string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "app_user_test" {
	id   = "%s"
	name = "%s"
}

resource "humanitec_resource_application_user" "another_user" {
  app_id  = humanitec_application.app_user_test.id
  user_id = "%s"
  role    = "%s"
}
`, id, id, userID, role)
}
