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

func TestAccResourceSecretStore_AzureKV(t *testing.T) {
	id := fmt.Sprintf("azurekv-test-%d", time.Now().UnixNano())
	newId := fmt.Sprintf("azurekv-test-new-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecretStoreAzureKV(id, "tenant-id", "azurekv-url", "client-id", "client-secret"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.tenant_id", "tenant-id"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.url", "azurekv-url"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.auth.client_id", "client-id"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_secretstore.secret_store_azurekv_test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return id, nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"azurekv.auth"},
			},
			// Update and Read testing
			{
				Config: testAccSecretStoreAzureKV(id, "tenant-id", "azurekv-url-changed", "client-id-changed", "client-secret"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.tenant_id", "tenant-id"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.url", "azurekv-url-changed"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.auth.client_id", "client-id-changed"),
				),
			},
			// Replace and Read testing
			{
				Config: testAccSecretStoreAzureKV(newId, "tenant-id", "azurekv-url-changed", "client-id-changed", "client-secret"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "id", newId),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.tenant_id", "tenant-id"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.url", "azurekv-url-changed"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_azurekv_test", "azurekv.auth.client_id", "client-id-changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceSecretStore_Aws(t *testing.T) {
	id := fmt.Sprintf("awssm-test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecretStoreAwsSM(id, "access-key-id", "secret-access-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_awssm_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_awssm_test", "awssm.region", "eu-central-1"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_awssm_test", "awssm.auth.access_key_id", "access-key-id"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_secretstore.secret_store_awssm_test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return id, nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"awssm.auth"},
			},
			// Update and Read testing
			{
				Config: testAccSecretStoreAwsSM(id, "access-key-id-changed", "secret-access-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_awssm_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_awssm_test", "awssm.region", "eu-central-1"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_awssm_test", "awssm.auth.access_key_id", "access-key-id-changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceSecretStore_GcpSM(t *testing.T) {
	id := fmt.Sprintf("gcpsm-test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecretStoreGcpSM(id, "gcp-project", "secret-access-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_gcpsm_test", "gcpsm.project_id", "gcp-project"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_gcpsm_test", "gcpsm.auth.secret_access_key", "secret-access-key"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_secretstore.secret_store_gcpsm_test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return id, nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"gcpsm.auth"},
			},
			// Update and Read testing
			{
				Config: testAccSecretStoreGcpSM(id, "gcp-project-changed", "secret-access-key-changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_gcpsm_test", "gcpsm.project_id", "gcp-project-changed"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_gcpsm_test", "gcpsm.auth.secret_access_key", "secret-access-key-changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceSecretStore_Vault(t *testing.T) {
	id := fmt.Sprintf("vault-test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecretStoreVault(id, "vault-url", "vault-token", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.url", "vault-url"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.auth.token", "vault-token"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_secretstore.secret_store_vault_test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return id, nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vault.auth"},
			},
			// Update and Read testing
			{
				Config: testAccSecretStoreVault(id, "vault-url-changed", "vault-token-changed", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.url", "vault-url-changed"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.auth.token", "vault-token-changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceSecretStore_Vault_RemoveAuth(t *testing.T) {
	id := fmt.Sprintf("vault-test-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecretStoreVault(id, "vault-url", "vault-token", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.url", "vault-url"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.auth.token", "vault-token"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_secretstore.secret_store_vault_test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return id, nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vault.auth"},
			},
			// Update and Read testing
			{
				Config: testAccSecretStoreVaultNoAuth(id, "vault-url", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "primary", "false"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.url", "vault-url"),
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.auth.%", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceSecretStore_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	id := fmt.Sprintf("secret-store-%d", time.Now().UnixNano())

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
				Config: testAccSecretStoreVaultNoAuth(id, "dumburl", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_secretstore.secret_store_vault_test", "vault.url", "dumburl"),
					func(_ *terraform.State) error {
						// Manually delete the secret store via the API
						resp, err := client.DeleteOrgsOrgIdSecretstoresStoreIdWithResponse(ctx, orgID, id)
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

func testAccSecretStoreAzureKV(storeID, tenantID, url, clientID, clientSecret string) string {
	return fmt.Sprintf(`
	resource "humanitec_secretstore" "secret_store_azurekv_test" {
		id      = "%s"
		azurekv = {
			tenant_id   = "%s"
			url         = "%s"
			auth = {
				client_id     = "%s"
				client_secret = "%s"
			}
		}
	}
`, storeID, tenantID, url, clientID, clientSecret)
}

func testAccSecretStoreAwsSM(storeID, accessKeyID, secretAccessKey string) string {
	return fmt.Sprintf(`
	resource "humanitec_secretstore" "secret_store_awssm_test" {
		id      = "%s"
		awssm = {
			region   = "eu-central-1"
			auth = {
				access_key_id     = "%s"
				secret_access_key = "%s"
			}
		}
	}
`, storeID, accessKeyID, secretAccessKey)
}

func testAccSecretStoreGcpSM(storeID, projectID, secretAccessKey string) string {
	return fmt.Sprintf(`
	resource "humanitec_secretstore" "secret_store_gcpsm_test" {
		id = "%s"
		gcpsm = {
			project_id   = "%s"
			auth = {
				secret_access_key = "%s"
			}
		}
	}
`, storeID, projectID, secretAccessKey)
}

func testAccSecretStoreVault(storeID, url, token string, primary bool) string {
	return fmt.Sprintf(`
	resource "humanitec_secretstore" "secret_store_vault_test" {
		id      = "%s"
		primary = %v
		vault = {
			url  = "%s"
			auth = {
				token = "%s"
			}
		}
	}
`, storeID, primary, url, token)
}

func testAccSecretStoreVaultNoAuth(storeID, url string, primary bool) string {
	return fmt.Sprintf(`
	resource "humanitec_secretstore" "secret_store_vault_test" {
		id      = "%s"
		primary = %v
		vault = {
			url  = "%s"
		}
	}
`, storeID, primary, url)
}
