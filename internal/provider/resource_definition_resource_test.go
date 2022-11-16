package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDefinitionS3Resource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceDefinitionS3Resource("us-east-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "id", "s3-test"),
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "driver_inputs.values.region", "us-east-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_resource_definition.s3_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceDefinitionS3Resource("us-east-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "driver_inputs.values.region", "us-east-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceDefinitionS3Resource(region string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test" {
  id          = "s3-test"
  name        = "s3-test"
  type        = "s3"
  driver_type = "humanitec/s3"

  driver_inputs = {
    values = {
      "region" = "%s"
    }
  }
}
`, region)
}

func TestAccResourceDefinitionPostgresResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceDefinitionPostgresResource("test-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.postgres_test", "id", "postgres-test"),
					resource.TestCheckResourceAttr("humanitec_resource_definition.postgres_test", "driver_inputs.values.name", "test-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_resource_definition.postgres_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"driver_inputs.secrets"},
			},
			// Update and Read testing
			{
				Config: testAccResourceDefinitionPostgresResource("test-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.postgres_test", "driver_inputs.values.name", "test-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceDefinitionPostgresResource(name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "postgres_test" {
  id          = "postgres-test"
  name        = "postgres-test"
  type        = "postgres"
  driver_type = "humanitec/postgres-cloudsql-static"

  driver_inputs = {
    values = {
      "instance" = "test:test:test"
      "name" = "%s"
      "host" = "127.0.0.1"
      "port" = "5432"
    }
		secrets = {
      "username" = "test"
      "password" = "test"
    }
  }
}
`, name)
}
