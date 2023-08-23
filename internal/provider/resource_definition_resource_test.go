package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceDefinition(t *testing.T) {
	tests := []struct {
		name                         string
		configCreate                 func() string
		configUpdate                 func() string
		resourceAttrName             string
		resourceAttrNameIDValue      string
		resourceAttrNameUpdateKey    string
		resourceAttrNameUpdateValue1 string
		resourceAttrNameUpdateValue2 string
	}{
		{
			name: "S3",
			configCreate: func() string {
				return testAccResourceDefinitionS3Resource("us-east-1")
			},
			resourceAttrNameIDValue:      "s3-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values.region",
			resourceAttrNameUpdateValue1: "us-east-1",
			resourceAttrName:             "humanitec_resource_definition.s3_test",
			configUpdate: func() string {
				return testAccResourceDefinitionS3Resource("us-east-2")
			},
			resourceAttrNameUpdateValue2: "us-east-2",
		},
		{
			name: "Postgres",
			configCreate: func() string {
				return testAccResourceDefinitionPostgresResource("test-1")
			},
			resourceAttrNameIDValue:      "postgres-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values.name",
			resourceAttrNameUpdateValue1: "test-1",
			resourceAttrName:             "humanitec_resource_definition.postgres_test",
			configUpdate: func() string {
				return testAccResourceDefinitionPostgresResource("test-2")
			},
			resourceAttrNameUpdateValue2: "test-2",
		},
		{
			name: "GKE",
			configCreate: func() string {
				return testAccResourceDefinitionGKEResource("test-1")
			},
			resourceAttrNameIDValue:      "gke-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values.name",
			resourceAttrNameUpdateValue1: "test-1",
			resourceAttrName:             "humanitec_resource_definition.gke_test",
			configUpdate: func() string {
				return testAccResourceDefinitionGKEResource("test-2")
			},
			resourceAttrNameUpdateValue2: "test-2",
		},
		{
			name: "DNS",
			configCreate: func() string {
				return testAccResourceDefinitionDNSStaticResource("test-1")
			},
			resourceAttrNameIDValue:      "dns-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values.host",
			resourceAttrNameUpdateValue1: "test-1",
			resourceAttrName:             "humanitec_resource_definition.dns_test",
			configUpdate: func() string {
				return testAccResourceDefinitionDNSStaticResource("test-2")
			},
			resourceAttrNameUpdateValue2: "test-2",
		},
		{
			name: "Ingress",
			configCreate: func() string {
				return testAccResourceDefinitionIngressResource("test-1")
			},
			resourceAttrNameIDValue:      "ingress-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values.labels",
			resourceAttrNameUpdateValue1: "{\"name\":\"test-1\"}",
			resourceAttrName:             "humanitec_resource_definition.ingress_test",
			configUpdate: func() string {
				return testAccResourceDefinitionIngressResource("test-2")
			},
			resourceAttrNameUpdateValue2: "{\"name\":\"test-2\"}",
		},
		{
			name: "Provision",
			configCreate: func() string {
				return testAccResourceDefinitionProvisionResource("true")
			},
			resourceAttrNameIDValue:      "provision-test",
			resourceAttrNameUpdateKey:    "provision.awspolicy.match_dependents",
			resourceAttrNameUpdateValue1: "true",
			resourceAttrName:             "humanitec_resource_definition.provision_test",
			configUpdate: func() string {
				return testAccResourceDefinitionProvisionResource("false")
			},
			resourceAttrNameUpdateValue2: "false",
		},
		{
			name: "k8s-logging",
			configCreate: func() string {
				return testAccResourceDefinitionK8sLoggingResource("test-1")
			},
			resourceAttrNameIDValue:      "k8s-logging-test",
			resourceAttrNameUpdateKey:    "name",
			resourceAttrNameUpdateValue1: "test-1",
			resourceAttrName:             "humanitec_resource_definition.k8s_logging_test",
			configUpdate: func() string {
				return testAccResourceDefinitionK8sLoggingResource("test-2")
			},
			resourceAttrNameUpdateValue2: "test-2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: tc.configCreate(),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(tc.resourceAttrName, "id", tc.resourceAttrNameIDValue),
							resource.TestCheckResourceAttr(tc.resourceAttrName, tc.resourceAttrNameUpdateKey, tc.resourceAttrNameUpdateValue1),
						),
					},
					// ImportState testing
					{
						ResourceName:            tc.resourceAttrName,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"driver_inputs.secrets", "force_delete"},
					},
					// Update and Read testing
					{
						Config: tc.configUpdate(),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(tc.resourceAttrName, tc.resourceAttrNameUpdateKey, tc.resourceAttrNameUpdateValue2),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceDefinitionS3Resource(region string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test" {
  id          = "s3-test"
  name        = "s3-test"
  type        = "s3"
  driver_type = "humanitec/s3"

  driver_inputs = {
    values = {
      "region" = "%s"
    }
  }
}
`, region)
}

func testAccResourceDefinitionPostgresResource(name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "postgres_test" {
  id          = "postgres-test"
  name        = "postgres-test"
  type        = "postgres"
  driver_type = "humanitec/postgres-cloudsql-static"

  driver_inputs = {
    values = {
      "instance" = "test:test:test"
      "name" = "%s"
      "host" = "127.0.0.1"
      "port" = "5432"
    }
    secrets = {
      "username" = "test"
      "password" = "test"
    }
  }
}
`, name)
}

func testAccResourceDefinitionGKEResource(name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "gke_test" {
  id          = "gke-test"
  name        = "gke-test"
  type        = "k8s-cluster"
  driver_type = "humanitec/k8s-cluster-gke"

  driver_inputs = {
    values = {
      "loadbalancer" = "1.1.1.1"
      "name" = "%s"
      "project_id" = "test"
      "zone" = "europe-west3"
    }
    secrets = {
      "credentials" = "{}"
    }
  }
}
`, name)
}

func testAccResourceDefinitionDNSStaticResource(host string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "dns_test" {
  id          = "dns-test"
  name        = "dns-test"
  type        = "dns"
  driver_type = "humanitec/static"

  driver_inputs = {
    values = {
      host = "%s"
    }
  }
}
`, host)
}

func testAccResourceDefinitionIngressResource(host string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "ingress_test" {
  id          = "ingress-test"
  name        = "ingress-test"
  type        = "ingress"
  driver_type = "humanitec/ingress"

  driver_inputs = {
    values = {
      labels = jsonencode({
				name = "%s"
			})
			no_tls      = true
    }
  }
}
`, host)
}

func testAccResourceDefinitionK8sLoggingResource(name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "k8s_logging_test" {
  id          = "k8s-logging-test"
  name        = "%s"
  type        = "logging"
  driver_type = "humanitec/logging-k8s"
}
`, name)
}

func testAccResourceDefinitionProvisionResource(matchDependents string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "provision_test" {
	id          = "provision-test"
	name        = "provision-test"
	type        = "s3"
	driver_type = "humanitec/s3"
	provision = {
		"awspolicy" = {
			is_dependent = true
			match_dependents = %s
		}
	}

	driver_inputs = {
		values = {
			"region" = "us-east-1"
		}
	}
}
`, matchDependents)
}

func TestAccResourceDefinitionWithDefinition(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceDefinitionS3ResourceCriteria(`{
					app_id = "app1"
					env_id = "dev"
				}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "id", "s3-test-with-criteria"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "humanitec_resource_definition.s3_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"driver_inputs.secrets", "force_delete"},
			},
			// Update and Read testing
			{
				Config: testAccResourceDefinitionS3ResourceCriteria(`{
					app_id = "app1"
					env_id = "dev"
				}, {
					app_id = "app1"
					env_id = "stg"
				}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "id", "s3-test-with-criteria"),
				),
			},
			// Update and Read testing
			{
				Config: testAccResourceDefinitionS3ResourceCriteria(`{
					app_id = "app1"
					env_id = "stg"
				}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test", "id", "s3-test-with-criteria"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourceDefinitionS3ResourceCriteria(criteria string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test" {
  id          = "s3-test-with-criteria"
  name        = "s3-test"
  type        = "s3"
  driver_type = "humanitec/s3"

  driver_inputs = {
    values = {
      "region" = "us-east-1"
    }
  }

	criteria = [%s]
}
`, criteria)
}

func TestDiffCriteria(t *testing.T) {
	tests := []struct {
		name            string
		previous        *[]DefinitionResourceCriteriaModel
		current         *[]DefinitionResourceCriteriaModel
		expectedAdded   []DefinitionResourceCriteriaModel
		expectedRemoved []DefinitionResourceCriteriaModel
	}{
		{
			name:            "both_empty",
			expectedAdded:   []DefinitionResourceCriteriaModel{},
			expectedRemoved: []DefinitionResourceCriteriaModel{},
		},
		{
			name: "previous_empty",
			current: &[]DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app")},
			},
			expectedAdded: []DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app")},
			},
			expectedRemoved: []DefinitionResourceCriteriaModel{},
		},
		{
			name: "current_empty",
			previous: &[]DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app")},
			},
			expectedAdded: []DefinitionResourceCriteriaModel{},
			expectedRemoved: []DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app")},
			},
		},
		{
			name: "diff",
			previous: &[]DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app")},
				{AppID: types.StringValue("test-app"), EnvType: types.StringValue("test-env-type")},
			},
			current: &[]DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app")},
				{AppID: types.StringValue("test-app"), EnvID: types.StringValue("test-env")},
			},
			expectedAdded: []DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app"), EnvID: types.StringValue("test-env")},
			},
			expectedRemoved: []DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("test-app"), EnvType: types.StringValue("test-env-type")},
			},
		},
		{
			name: "diff_single_removed",
			previous: &[]DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("app1"), EnvID: types.StringValue("dev")},
				{AppID: types.StringValue("app1"), EnvID: types.StringValue("stg")},
			},
			current: &[]DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("app1"), EnvID: types.StringValue("stg")},
			},
			expectedAdded: []DefinitionResourceCriteriaModel{},
			expectedRemoved: []DefinitionResourceCriteriaModel{
				{AppID: types.StringValue("app1"), EnvID: types.StringValue("dev")},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			added, removed := diffCriteria(tc.previous, tc.current)
			assert.Equal(tc.expectedAdded, added)
			assert.Equal(tc.expectedRemoved, removed)
		})
	}
}

func TestDriverInputToMap(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	input, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"stringValue":     "test",
		"integerValue":    "1",
		"refIntegerValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"objectValue":     "{\"nested\":\"value\"}",
		"refObjectValue":  "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"booleanValue":    "true",
		"refBooleanValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
	})
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"properties": map[string]interface{}{
					"stringValue": map[string]interface{}{
						"type": "string",
					},
					"integerValue": map[string]interface{}{
						"type": "integer",
					},
					"refIntegerValue": map[string]interface{}{
						"type": "integer",
					},
					"objectValue": map[string]interface{}{
						"type": "object",
					},
					"refObjectValue": map[string]interface{}{
						"type": "object",
					},
					"booleanValue": map[string]interface{}{
						"type": "boolean",
					},
					"refBooleanValue": map[string]interface{}{
						"type": "boolean",
					},
				},
			},
		},
	}

	m, diags := driverInputToMap(ctx, input, schema, "test")
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}

	expected := map[string]interface{}{
		"stringValue":     "test",
		"integerValue":    1,
		"refIntegerValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"objectValue": map[string]interface{}{
			"nested": "value",
		},
		"refObjectValue":  "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"booleanValue":    true,
		"refBooleanValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
	}
	assert.Equal(expected, m)
}

func TestDriverInputToMap_MissingValue(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	input, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"missingValue": "test",
	})
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"properties": map[string]interface{}{},
			},
		},
	}

	_, diags = driverInputToMap(ctx, input, schema, "test")
	assert.True(diags.HasError())
}

func TestDriverInputToMap_UnexpectedType(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	input, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"unexpectedValue": "test",
	})
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"properties": map[string]interface{}{
					"unexpectedValue": map[string]interface{}{
						"type": "unexpected",
					},
				},
			},
		},
	}

	_, diags = driverInputToMap(ctx, input, schema, "test")
	assert.True(diags.HasError())
}

func TestParseMapInput(t *testing.T) {
	assert := assert.New(t)

	input := map[string]interface{}{
		"stringValue":     "test",
		"integerValue":    float64(1),
		"refIntegerValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"objectValue": map[string]interface{}{
			"nested": "value",
		},
		"refObjectValue":  "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"booleanValue":    true,
		"refBooleanValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
	}
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"properties": map[string]interface{}{
					"stringValue": map[string]interface{}{
						"type": "string",
					},
					"integerValue": map[string]interface{}{
						"type": "integer",
					},
					"refIntegerValue": map[string]interface{}{
						"type": "integer",
					},
					"objectValue": map[string]interface{}{
						"type": "object",
					},
					"refObjectValue": map[string]interface{}{
						"type": "object",
					},
					"booleanValue": map[string]interface{}{
						"type": "boolean",
					},
					"refBooleanValue": map[string]interface{}{
						"type": "boolean",
					},
				},
			},
		},
	}

	m, diags := parseMapInput(input, schema, "test")
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}

	expected := map[string]string{
		"stringValue":     "test",
		"integerValue":    "1",
		"refIntegerValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"objectValue":     "{\"nested\":\"value\"}",
		"refObjectValue":  "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
		"booleanValue":    "true",
		"refBooleanValue": "${resources.k8s-cluster#k8s-cluster.outputs.credentials}",
	}
	assert.Equal(expected, m)
}

func TestParseMapInput_MissingValue(t *testing.T) {
	assert := assert.New(t)

	input := map[string]interface{}{
		"missingValue": "test",
	}
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"properties": map[string]interface{}{},
			},
		},
	}

	_, diags := parseMapInput(input, schema, "test")
	assert.True(diags.HasError())
}

func TestParseMapInput_UnexpectedType(t *testing.T) {
	assert := assert.New(t)

	input := map[string]interface{}{
		"unexpectedValue": "test",
	}
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"properties": map[string]interface{}{
					"unexpectedValue": map[string]interface{}{
						"type": "unexpected",
					},
				},
			},
		},
	}

	_, diags := parseMapInput(input, schema, "test")
	assert.True(diags.HasError())
}
