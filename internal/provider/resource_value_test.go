package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceValue(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())
	key := "VAL_1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValue(appID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "description", "Example value"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_val1",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", appID, key), nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValue(appID, key, "Example value changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "description", "Example value changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceValueWithEnv(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-env-%d", time.Now().UnixNano())
	envID := "dev"
	key := "VAL_1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "description", "Example value"),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "env_id", "dev"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_env_val1",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", appID, envID, key), nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, "Example value changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "description", "Example value changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceVALUETestAccResourceValue(appID, key, description string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"
}

resource "humanitec_value" "app_val1" {
	app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "%s"
	value       = "TEST"
	is_secret   = false
}
`, appID, key, description)
}

func testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, description string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"

	env = {
		id   = "%s"
		name = "dev"
		type = "development"
	}
}

resource "humanitec_value" "app_val1" {
	app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "app value"
	value       = "TEST"
	is_secret   = false
}

resource "humanitec_value" "app_env_val1" {
	app_id = humanitec_application.val_test.id
	env_id = "%s"

	key         = "%s"
	description = "%s"
	value       = "TEST"
	is_secret   = false

	depends_on = [
		humanitec_value.app_val1
	]
}
`, appID, envID, key, envID, key, description)
}
