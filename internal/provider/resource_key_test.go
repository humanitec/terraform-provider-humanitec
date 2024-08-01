package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceKeys(t *testing.T) {
	key := getPublicKey(t)
	var id string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceKey(key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_key.key_test", "key", key),
					resource.TestCheckResourceAttrSet("humanitec_key.key_test", "id"),
					resource.TestCheckResourceAttrSet("humanitec_key.key_test", "fingerprint"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_key.key_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id = s.RootModule().Resources["humanitec_key.key_test"].Primary.Attributes["id"]
					return id, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_key.key_test", "key", key),
					resource.TestCheckResourceAttr("humanitec_key.key_test", "id", id),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceKey_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	orgID := os.Getenv("HUMANITEC_ORG")
	token := os.Getenv("HUMANITEC_TOKEN")
	apiHost := os.Getenv("HUMANITEC_HOST")
	if apiHost == "" {
		apiHost = humanitec.DefaultAPIHost
	}

	var client *humanitec.Client
	var err error

	key := getPublicKey(t)
	var id string

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
				Config: testAccResourceKey(key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_key.key_test", "key", key),
					resource.TestCheckResourceAttrSet("humanitec_key.key_test", "id"),
					resource.TestCheckResourceAttrSet("humanitec_key.key_test", "fingerprint"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_key.key_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					id = s.RootModule().Resources["humanitec_key.key_test"].Primary.Attributes["id"]
					return id, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_key.key_test", "key", key),
					resource.TestCheckResourceAttr("humanitec_key.key_test", "id", id),
					func(_ *terraform.State) error {
						// Manually delete the public key via the API
						resp, err := client.DeletePublicKeyWithResponse(ctx, orgID, id)
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
