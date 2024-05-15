package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceEnvironment(t *testing.T) {
	id := fmt.Sprintf("env-%d", time.Now().Second())
	appID := fmt.Sprintf("app-%d", time.Now().UnixNano())
	name := "Env Name"
	updatedName := "New Env Name"
	envType := "development"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCreateResourceEnvironment(appID, id, name, envType, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "app_id", appID),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "name", name),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "type", envType),
				),
			},
			// Update testing
			{
				Config: testAccCreateResourceEnvironment(appID, id, updatedName, envType, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "app_id", appID),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "name", updatedName),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "type", envType),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_environment.env_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", appID, id), nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "app_id", appID),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "name", updatedName),
					resource.TestCheckResourceAttr("humanitec_environment.env_test", "type", envType),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCreateResourceEnvironment(appID, id, name, envType, fromDeployID string) string {
	fromDeployIDLine := ""
	if fromDeployID != "" {
		fromDeployIDLine = fmt.Sprintf(`from_deploy_id = "%s"`, fromDeployID)
	}

	return fmt.Sprintf(`
	resource "humanitec_application" "app_test" {
		id          = "%s"
		name        = "test-app"
	}

	resource "humanitec_environment" "env_test" {
		app_id = humanitec_application.app_test.id
		id     = "%s"
		name   = "%s"
		type   = "%s"
		%s
	}
`, appID, id, name, envType, fromDeployIDLine)
}
