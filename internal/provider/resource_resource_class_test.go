package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceClass(t *testing.T) {
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())
	description := "test-description"
	updatedDescription := "test-updated-description"
	resourceType := "mysql"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceClass(id, description, resourceType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "description", description),
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "resource_type", resourceType),
				),
			},
			// Update testing
			{
				Config: testAccResourceClass(id, updatedDescription, resourceType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "description", updatedDescription),
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "resource_type", resourceType),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_resource_class.class_test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", resourceType, id), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceClass(id, description, resourceType string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_class" "class_test" {
  id            = "%s"
  description   = "%s"
  resource_type = "%s"
}
`, id, description, resourceType)
}
