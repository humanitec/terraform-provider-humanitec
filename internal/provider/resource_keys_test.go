package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceKeys(t *testing.T) {
	key := getPublicKey(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceKey(key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_key.key_test", "key", key),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_key.key_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id := s.RootModule().Resources["humanitec_key.key_test"].Primary.Attributes["id"]
					return id, nil
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceKey(key string) string {
	return fmt.Sprintf(`
	resource "humanitec_key" "key_test" {
		key = %v
	}
	
	output "key_id" {
		value = humanitec_key.key_test.id
	}
`, toSingleLineTerraformString(key))
}
