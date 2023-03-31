package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceApplication(t *testing.T) {
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceApplication(id, "test-app-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application.app_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_application.app_test", "name", "test-app-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_application.app_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceApplication(id, name string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "app_test" {
  id          = "%s"
  name        = "%s"
}
`, id, name)
}

func TestAccResourceApplicationWithInitialEnv(t *testing.T) {
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceApplicationWithEnv(id, "test-app-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application.app_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_application.app_test", "name", "test-app-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_application.app_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"env"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceApplicationWithEnv(id, name string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "app_test" {
  id          = "%s"
  name        = "%s"

	env = {
		name = "test"
		id   = "test"
		type = "development"
	}
}
`, id, name)
}
