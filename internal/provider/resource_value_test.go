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

func TestAccResourceValue(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())
	key := "VAL_1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValue(appID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "description", "Example value"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_val1",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", appID, key), nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValue(appID, key, "Example value changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "description", "Example value changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceValueWithSecretValue(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())
	key := "VAL_SECRET_1"
	orgID := os.Getenv("HUMANITEC_ORG")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueSecret(appID, key, "Example value with secret"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", fmt.Sprintf("orgs/%s/apps/%s/secret_values/%s/.value", orgID, appID, key)),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.version", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_val_with_secret",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", appID, key), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_ref", "value"},
			},
			// Update
			{
				Config: testAccResourceVALUETestAccResourceValueSecret(appID, key, "Example value with secret changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret changed"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", fmt.Sprintf("orgs/%s/apps/%s/secret_values/%s/.value", orgID, appID, key)),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.version", "2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceValueWithSecretValueSecretRefValue(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())
	key := "VAL_SECRET_REF_VALUE_1"
	orgID := os.Getenv("HUMANITEC_ORG")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueSecretRefValue(appID, key, "Example value with secret set via secret reference value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret set via secret reference value"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", fmt.Sprintf("orgs/%s/apps/%s/secret_values/%s/.value", orgID, appID, key)),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.version", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_val_with_secret",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", appID, key), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_ref"},
			},
			{
				Config: testAccResourceVALUETestAccResourceValueSecretRefValue(appID, key, "Example value with secret set via secret reference value changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret set via secret reference value changed"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", fmt.Sprintf("orgs/%s/apps/%s/secret_values/%s/.value", orgID, appID, key)),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.version", "2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceValueWithSecretRef(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())
	key := "VAL_SECRET_REF_1"
	orgID := os.Getenv("HUMANITEC_ORG")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueSecretRef(appID, key, "path/to/secret", "Example value with secret reference", "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret reference"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", "path/to/secret"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.version", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_val_with_secret",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", appID, key), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_ref"},
			},
			// Update and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueSecretRef(appID, key, "path/to/secret/changed", "Example value with secret reference changed", "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret reference changed"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", "path/to/secret/changed"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.version", "2"),
				),
			},
			// Update and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueSecret(appID, key, "Example value with secret reference updated with plain value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "description", "Example value with secret reference updated with plain value"),
					resource.TestCheckResourceAttr("humanitec_value.app_val_with_secret", "secret_ref.ref", fmt.Sprintf("orgs/%s/apps/%s/secret_values/%s/.value", orgID, appID, key)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceValueDeletedOutManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())

	key := "VAL_1"

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
				Config: testAccResourceVALUETestAccResourceValue(appID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "description", "Example value"),
					func(_ *terraform.State) error {
						// Manually delete the value via the API
						resp, err := client.DeleteOrgsOrgIdAppsAppIdValuesKeyWithResponse(ctx, orgID, appID, key)
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

func TestAccResourceValueWithEnv(t *testing.T) {
	appID := fmt.Sprintf("val-test-app-env-%d", time.Now().UnixNano())

	envID := "dev"
	key := "VAL_1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "description", "Example value"),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "env_id", "dev"),
				),
			},
			// ImportState testing
			{
				ResourceName: "humanitec_value.app_env_val1",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", appID, envID, key), nil
				},
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, "Example value changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "description", "Example value changed"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccResourceValueWithEnvEnvDeletedOutManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	appID := fmt.Sprintf("val-test-app-%d", time.Now().UnixNano())

	envID := "dev"
	key := "VAL_1"

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
				Config: testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "description", "Example value"),
					func(_ *terraform.State) error {
						// Manually delete the env via the API
						resp, err := client.DeleteEnvironmentWithResponse(ctx, orgID, appID, envID)
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
			// Reapply after manually deleted testing
			{
				Config: testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, "Example value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_value.app_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "key", key),
					resource.TestCheckResourceAttr("humanitec_value.app_env_val1", "description", "Example value"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceVALUETestAccResourceValue(appID, key, description string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"
}

resource "humanitec_value" "app_val1" {
	app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "%s"
	value       = "TEST"
	is_secret   = false
}
`, appID, key, description)
}

func testAccResourceVALUETestAccResourceValueWithEnv(appID, envID, key, description string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"
}

resource "humanitec_environment" "dev" {
	app_id = humanitec_application.val_test.id
	id = "%s"
	name = "dev"
	type = "development"
}

resource "humanitec_value" "app_val1" {
	app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "app value"
	value       = "TEST"
	is_secret   = false
}

resource "humanitec_value" "app_env_val1" {
	app_id = humanitec_application.val_test.id
	env_id = humanitec_environment.dev.id

	key         = "%s"
	description = "%s"
	value       = "TEST"
	is_secret   = false

	depends_on = [
		humanitec_value.app_val1
	]
}
`, appID, envID, key, key, description)
}

func testAccResourceVALUETestAccResourceValueSecretRef(appID, key, secretPath, description, version string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"
}

resource "humanitec_value" "app_val_with_secret" {
	app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "%s"
  is_secret   = true
  secret_ref  = {
	ref     = "%s"
	store   = "external-store-id"
	version = "%s"
  }
}
`, appID, key, description, secretPath, version)
}

func testAccResourceVALUETestAccResourceValueSecret(appID, key, description string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"
}

resource "humanitec_value" "app_val_with_secret" {
  app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "%s"
  is_secret   = true
  value       = "secret"
}
`, appID, key, description)
}

func testAccResourceVALUETestAccResourceValueSecretRefValue(appID, key, description string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "val_test" {
	id   = "%s"
	name = "val-test"
}

resource "humanitec_value" "app_val_with_secret" {
  app_id = humanitec_application.val_test.id

  key         = "%s"
  description = "%s"
  is_secret   = true
  secret_ref  = {
	  value = "secret"
  }
}
`, appID, key, description)
}
