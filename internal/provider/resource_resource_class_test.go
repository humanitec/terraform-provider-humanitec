package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/stretchr/testify/assert"
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

func TestAccResourceClass_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("test-class-%d", time.Now().UnixNano())
	description := "test-description"
	resourceType := "mysql"

	orgID := os.Getenv("HUMANITEC_ORG")
	token := os.Getenv("HUMANITEC_TOKEN")
	apiHost := os.Getenv("HUMANITEC_HOST")
	if apiHost == "" {
		apiHost = humanitec.DefaultAPIHost
	}

	var client *humanitec.Client
	var err error

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			client, err = NewHumanitecClient(apiHost, token, "test", nil)
			assert.NoError(err)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClass(id, description, resourceType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_class.class_test", "id", id),
					func(_ *terraform.State) error {
						// Manually delete the resource class via the API
						resp, err := client.DeleteResourceClassWithResponse(ctx, orgID, resourceType, id)
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

func testAccResourceClass(id, description, resourceType string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_class" "class_test" {
  id            = "%s"
  description   = "%s"
  resource_type = "%s"
}
`, id, description, resourceType)
}
