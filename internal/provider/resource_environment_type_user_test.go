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

func TestAccResourceEnvironmentTypeUser(t *testing.T) {
	id := fmt.Sprintf("env-type-user-test-%d", time.Now().UnixNano())
	testUserID := "c0725726-0613-43d4-8398-907d07fba2e4"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "role", "deployer"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_environment_type_user.another_user",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", id, testUserID), nil
				},
			},
			// Update and Read testing
			{
				// At the moment, there is nothing we can update :-/
				Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "role", "deployer"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceEnvironmentTypeUserDeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("env-type-user-test-%d", time.Now().UnixNano())
	testUserID := "c0725726-0613-43d4-8398-907d07fba2e4"

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
				Config: testAccResourceEnvironmentTypeUser(id, testUserID, "deployer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "id", fmt.Sprintf("%s/%s", id, testUserID)),
					resource.TestCheckResourceAttr("humanitec_environment_type_user.another_user", "role", "deployer"),
					func(_ *terraform.State) error {
						// Manually delete the environment type via the API
						resp, err := client.DeleteEnvironmentTypeWithResponse(ctx, orgID, id)
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

func testAccResourceEnvironmentTypeUser(id, userID, role string) string {
	return fmt.Sprintf(`
resource "humanitec_environment_type" "qa" {
	id            = "%s"
	description   = "%s"
}

resource "humanitec_environment_type_user" "another_user" {
  env_type_id = humanitec_environment_type.qa.id
  user_id     = "%s"
  role        = "%s"
}
`, id, id, userID, role)
}
