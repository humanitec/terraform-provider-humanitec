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

func TestAccResourceApplicationDeletedOutManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())

	orgID := os.Getenv("HUMANITEC_ORG_ID")
	token := os.Getenv("HUMANITEC_TOKEN")

	var client *humanitec.Client
	var err error

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			client, err = NewHumanitecClient(humanitec.DefaultAPIHost, token, "test", nil)
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
					func(_ *terraform.State) error {
						// Manually delete the application via the API
						resp, err := client.DeleteOrgsOrgIdAppsAppIdWithResponse(ctx, orgID, id)
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
