package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceArtifactVersion(t *testing.T) {
	name := fmt.Sprintf("registry.humanitec.io/my-org/my-service-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceArtifactVersion(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_artefact_version.container_image", "name", name),
					resource.TestCheckResourceAttr("humanitec_artefact_version.container_image", "type", "container"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_artefact_version.container_image",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceArtifactVersion(name string) string {
	return fmt.Sprintf(`
resource "humanitec_artefact_version" "container_image" {
  type        = "container"
  name        = "%s"
}
`, name)
}

func TestAccResourceArtifactVersionWithOptional(t *testing.T) {
	name := fmt.Sprintf("registry.humanitec.io/my-org/my-service-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceArtifactVersionWithVersion(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_artefact_version.container_image", "name", name),
					resource.TestCheckResourceAttr("humanitec_artefact_version.container_image", "type", "container"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_artefact_version.container_image",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceArtifactVersionWithVersion(name string) string {
	return fmt.Sprintf(`
resource "humanitec_artefact_version" "container_image" {
  type        = "container"
  name        = "%s"
	version     = "1.2.3"
}
`, name)
}
