package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceRegistry(t *testing.T) {
	testCases := []struct {
		name         string
		configCreate func(id, registry string) string
		configUpdate func(id, registry string) string
	}{
		{
			name: "WithSecrets",
			configCreate: func(id, registry string) string {
				return testAccResourceRegistry(id, registry, false)
			},
			configUpdate: func(id, registry string) string {
				return testAccResourceRegistry(id, registry, true)
			},
		},
		{
			name: "WithCreds",
			configCreate: func(id, registry string) string {
				return testAccResourceRegistryCreds(id, registry, false)
			},
			configUpdate: func(id, registry string) string {
				return testAccResourceRegistryCreds(id, registry, true)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprintf("test-%d", time.Now().UnixNano())
			registry := fmt.Sprintf("test-%d.com.pl", time.Now().UnixNano())

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: tc.configCreate(id, registry),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_registry.registry_test", "id", id),
							resource.TestCheckResourceAttr("humanitec_registry.registry_test", "registry", registry),
							resource.TestCheckResourceAttr("humanitec_registry.registry_test", "enable_ci", "false"),
						),
					},
					// ImportState testing
					{
						ResourceName:            "humanitec_registry.registry_test",
						ImportStateId:           id,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"creds"},
					},
					// Update testing
					{
						Config: tc.configUpdate(id, registry),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_registry.registry_test", "id", id),
							resource.TestCheckResourceAttr("humanitec_registry.registry_test", "registry", registry),
							resource.TestCheckResourceAttr("humanitec_registry.registry_test", "enable_ci", "true"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceRegistry(id, registry string, enable_ci bool) string {
	return fmt.Sprintf(`
resource "humanitec_registry" "registry_test" {
	id     = "%s"
	registry = "%s"
	type = "secret_ref"
	enable_ci = %t
	secrets = {
		"cluster-a" = {
		namespace = "example-namespace"
		secret = "path/to/secret"
		},
		"cluster-b" = {
		namespace = "example-namespace"
		secret = "path/to/secret"
		}
	}
}`, id, registry, enable_ci)
}

func testAccResourceRegistryCreds(id, registry string, enable_ci bool) string {
	return fmt.Sprintf(`
resource "humanitec_registry" "registry_test" {
	id     = "%s"
	registry = "%s"
	type = "amazon_ecr"
	enable_ci = %t
	creds = {
		username = "test-username"
		password = "test-password"
	}
}`, id, registry, enable_ci)
}

func TestParseRegistryModel(t *testing.T) {
	assert := assert.New(t)

	registry := &RegistryModel{
		ID: types.StringValue("test-id"),
		Creds: &RegistryCredsModel{
			Username: types.StringValue("test-username"),
			Password: types.StringValue("test-password"),
		},
	}

	model, diags := parseRegistryModel(registry)
	assert.Empty(diags)
	assert.Equal("test-id", model.Id)
	assert.Equal("test-username", model.Creds.Username)
	assert.Equal("test-password", model.Creds.Password)
}
