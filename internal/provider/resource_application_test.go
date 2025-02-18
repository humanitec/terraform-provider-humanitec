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

func TestAccResourceApplication(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())

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
			// Create and Read testing
			{
				Config: testAccResourceApplication(id, "test-app-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application.app_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_application.app_test", "name", "test-app-1"),
					testCheckNoEnvironmentsAreCreated(ctx, &client, orgID, id),
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

func TestAccResourceApplicationDeletedOutManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())

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
			// Create and Read testing
			{
				Config: testAccResourceApplication(id, "test-app-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application.app_test", "id", id),
					resource.TestCheckResourceAttr("humanitec_application.app_test", "name", "test-app-1"),
					testCheckNoEnvironmentsAreCreated(ctx, &client, orgID, id),
					func(_ *terraform.State) error {
						// Manually delete the application via the API
						resp, err := client.DeleteApplicationWithResponse(ctx, orgID, id)
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

func testCheckNoEnvironmentsAreCreated(ctx context.Context, clientPtr **humanitec.Client, orgId, appId string) func(_ *terraform.State) error {
	return func(s *terraform.State) error {
		if clientPtr == nil {
			return fmt.Errorf("clientPtr is nil")
		}
		if *clientPtr == nil {
			return fmt.Errorf("client is nil")
		}
		client := *clientPtr
		resp, err := client.ListEnvironmentsWithResponse(ctx, orgId, appId)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("expected status code 200, got %d, body: %s", resp.StatusCode(), string(resp.Body))
		}

		if resp.JSON200 != nil && len(*resp.JSON200) != 0 {
			return fmt.Errorf("expected no environments to be created, got %d", len(*resp.JSON200))
		}

		return nil
	}
}
