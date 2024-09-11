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

func TestAccResourceAccountResource(t *testing.T) {
	id := fmt.Sprintf("aws-test-%d", time.Now().UnixNano())
	role := fmt.Sprintf("arn:aws:iam::0000000:role/test-role-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceAccountResource(id, "aws-test-1", role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_account.aws_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_resource_account.aws_test", "name", "aws-test-1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_resource_account.aws_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
			// Update and Read testing
			{
				Config: testAccResourceAccountResource(id, "aws-test-2", role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_account.aws_test", "name", "aws-test-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceAccountResource_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("aws-test-%d", time.Now().UnixNano())
	role := fmt.Sprintf("arn:aws:iam::0000000:role/test-role-%d", time.Now().UnixNano())

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
				Config: testAccResourceAccountResource(id, "aws-test-2", role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_account.aws_test", "id", id),
					func(_ *terraform.State) error {
						// Manually delete the resource account via the API
						resp, err := client.DeleteResourceAccountWithResponse(ctx, orgID, id)
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

func testAccResourceAccountResource(id, name, role string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_account" "aws_test" {
  id          = "%s"
  name        = "%s"
  type        = "aws-role"
  credentials = jsonencode({
   aws_role = "%s" 
   external_id = "test_external_id"
  })
}
`, id, name, role)
}
