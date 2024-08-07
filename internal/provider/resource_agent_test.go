package provider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/stretchr/testify/assert"
)

func TestAccAgent(t *testing.T) {
	id := fmt.Sprintf("agent-test-%d", time.Now().UnixNano())
	description := "Demo Agent"
	publicKeyOne := getPublicKey(t)
	publicKeyTwo := getPublicKey(t)
	publicKeyThree := getPublicKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCreateAgent(id, description, publicKeyOne, publicKeyTwo),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_agent.agent_test", "description", description),
					resource.TestCheckResourceAttr("humanitec_agent.agent_test", "public_keys.#", "2"),
					resource.TestCheckResourceAttrWith("humanitec_agent.agent_test", "public_keys.0.key", func(value string) error {
						if value != publicKeyOne && value != publicKeyTwo {
							return fmt.Errorf("unexpected value: %v", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("humanitec_agent.agent_test", "public_keys.1.key", func(value string) error {
						if value != publicKeyOne && value != publicKeyTwo {
							return fmt.Errorf("unexpected value: %v", value)
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_agent.agent_test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return id, nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCreateAgent(id, "", publicKeyOne, publicKeyThree),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_agent.agent_test", "description", ""),
					resource.TestCheckResourceAttr("humanitec_agent.agent_test", "public_keys.#", "2"),
					resource.TestCheckResourceAttrWith("humanitec_agent.agent_test", "public_keys.0.key", func(value string) error {
						if value != publicKeyOne && value != publicKeyThree {
							return fmt.Errorf("unexpected value: %v", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("humanitec_agent.agent_test", "public_keys.1.key", func(value string) error {
						if value != publicKeyOne && value != publicKeyThree {
							return fmt.Errorf("unexpected value: %v", value)
						}
						return nil
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceAgent_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("agent-test-%d", time.Now().UnixNano())

	orgID := os.Getenv("HUMANITEC_ORG")
	token := os.Getenv("HUMANITEC_TOKEN")
	apiHost := os.Getenv("HUMANITEC_HOST")
	if apiHost == "" {
		apiHost = humanitec.DefaultAPIHost
	}

	var client *humanitec.Client
	var err error
	publicKeyOne := getPublicKey(t)
	publicKeyTwo := getPublicKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			client, err = NewHumanitecClient(apiHost, token, "test", nil)
			assert.NoError(err)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateAgent(id, "my agent", publicKeyOne, publicKeyTwo),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_agent.agent_test", "public_keys.#", "2"),
					func(_ *terraform.State) error {
						// Manually delete the agent via the API
						resp, err := client.DeleteAgentWithResponse(ctx, orgID, id)
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

func testAccCreateAgent(id, description string, publicKey, otherPublicKey string) string {
	return fmt.Sprintf(`
	resource "humanitec_agent" "agent_test" {
		id      = "%s"
		description = "%s"
		public_keys = [
			{
				key = %v
			},
			{
				key = %v
			}
		]
	}
`, id, description, toSingleLineTerraformString(publicKey), toSingleLineTerraformString(otherPublicKey))
}

func getPublicKey(t *testing.T) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	assert.NoError(t, err)

	derBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	})

	return string(pem)
}

func toSingleLineTerraformString(s string) string {
	return fmt.Sprintf("%q", s)
}
