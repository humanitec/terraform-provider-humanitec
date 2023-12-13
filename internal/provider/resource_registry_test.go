package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRegistry(t *testing.T) {
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())
	registry := fmt.Sprintf("test-%d.com.pl", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceRegistry(id, registry, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_registry.registry_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_registry.registry_test", "registry", registry),
					resource.TestCheckResourceAttr("humanitec_registry.registry_test", "enable_ci", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_registry.registry_test",
				ImportStateId: id,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccResourceRegistry(id, registry, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_registry.registry_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_registry.registry_test", "registry", registry),
					resource.TestCheckResourceAttr("humanitec_registry.registry_test", "enable_ci", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceRegistry(id, registry string, enable_ci bool) string {
	return fmt.Sprintf(`
resource "humanitec_registry" "registry_test" {
	id     = "%s"
	registry = "%s"
	type = "secret_ref"
	enable_ci = %t
	secrets = {
		"cluster-a" = {
		namespace = "example-namespace"
		secret = "path/to/secret"
		},
		"cluster-b" = {
		namespace = "example-namespace"
		secret = "path/to/secret"
		}
	}
}`, id, registry, enable_ci)
}
