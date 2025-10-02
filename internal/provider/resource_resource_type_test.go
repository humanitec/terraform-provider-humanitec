package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceResourceType(t *testing.T) {
	orgID := os.Getenv("HUMANITEC_ORG")
	id := fmt.Sprintf("%s/test-type-%d", orgID, time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceResourceType(id, "test-name-1", "test-category-1", "direct", "{}", "{}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "id", id),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "name", "test-name-1"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "category", "test-category-1"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "use", "direct"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "inputs_schema", "{}"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "outputs_schema", "{}"),
				),
			},
			// Update and Read testing
			{
				Config: testAccResourceResourceType(id, "test-name-2", "test-category-2", "indirect", `{"a":1}`, `{"b":2}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "id", id),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "name", "test-name-2"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "category", "test-category-2"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "use", "indirect"),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "inputs_schema", `{"a":1}`),
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "outputs_schema", `{"b":2}`),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_resource_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceResourceType(id, name, category, use, inputsSchema, outputsSchema string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_type" "test" {
  id             = "%s"
  name           = "%s"
  category       = "%s"
  use            = "%s"
  inputs_schema  = jsonencode(%s)
  outputs_schema = jsonencode(%s)
}
`, id, name, category, use, inputsSchema, outputsSchema)
}

func TestAccResourceResourceType_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	orgID := os.Getenv("HUMANITEC_ORG")
	id := fmt.Sprintf("%s/test-type-%d", orgID, time.Now().UnixNano())

	token := os.Getenv("HUMANITEC_TOKEN")
	host := os.Getenv("HUMANITEC_HOST")
	if host == "" {
		host = humanitec.DefaultAPIHost
	}

	var client *humanitec.Client
	var err error

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			client, err = NewHumanitecClient(host, token, "test", nil)
			assert.NoError(err)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceResourceType(id, "test-name-1", "test-category-1", "direct", "{}", "{}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_type.test", "id", id),
					func(_ *terraform.State) error {
						// Manually delete the resource type via the API
						resp, err := client.DeleteResourceTypeWithResponse(ctx, orgID, id)
						if err != nil {
							return err
						}

						if resp.StatusCode() != 204 {
							return fmt.Errorf("expected status code 204, got %d, body: %s", resp.StatusCode(), string(resp.Body))
						}

						return nil
					},
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
