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

func TestAccResourceApplicationUser(t *testing.T) {
	id := fmt.Sprintf("app-user-test-%d", time.Now().UnixNano())
	testUserID := "1b305f15-f18f-4357-8311-01f88ed99d1b"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceApplicationUser(id, testUserID, "owner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application_user.another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
					resource.TestCheckResourceAttr("humanitec_application_user.another_user", "role", "owner"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_application_user.another_user",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", id, testUserID), nil
				},
			},
			// Update and Read testing
			{
				Config: testAccResourceApplicationUser(id, testUserID, "developer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application_user.another_user", "role", "developer"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceApplicationUserDeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("app-user-test-%d", time.Now().UnixNano())
	testUserID := "1b305f15-f18f-4357-8311-01f88ed99d1b"

	orgID := os.Getenv("HUMANITEC_ORG")
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
				Config: testAccResourceApplicationUser(id, testUserID, "owner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_application_user.another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
					resource.TestCheckResourceAttr("humanitec_application_user.another_user", "role", "owner"),
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

func testAccResourceApplicationUser(id, userID, role string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "app_user_test" {
	id   = "%s"
	name = "%s"
}

resource "humanitec_application_user" "another_user" {
  app_id  = humanitec_application.app_user_test.id
  user_id = "%s"
  role    = "%s"
}
`, id, id, userID, role)
}
