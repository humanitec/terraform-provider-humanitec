package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDefinition(t *testing.T) {
	timestamp := time.Now().UnixNano()
	tests := []struct {
		name                         string
		configCreate                 func() string
		configUpdate                 func() string
		resourceAttrName             string
		resourceAttrNameIDValue      string
		resourceAttrNameUpdateKey    string
		resourceAttrNameUpdateValue1 string
		resourceAttrNameUpdateValue2 string
		importStateVerifyIgnore      []string
	}{
		{
			name: "S3",
			configCreate: func() string {
				return testAccResourceDefinitionS3Resource(fmt.Sprintf("s3-test-%d", timestamp), "us-east-1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("s3-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"region\":\"us-east-1\"}",
			resourceAttrName:             "humanitec_resource_definition.s3_test",
			configUpdate: func() string {
				return testAccResourceDefinitionS3Resource(fmt.Sprintf("s3-test-%d", timestamp), "us-east-2")
			},
			resourceAttrNameUpdateValue2: "{\"region\":\"us-east-2\"}",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "Postgres",
			configCreate: func() string {
				return testAccResourceDefinitionPostgresResource(fmt.Sprintf("postgres-test-%d", timestamp), "test-1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("postgres-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"host\":\"127.0.0.1\",\"instance\":\"test:test:test\",\"name\":\"test-1\",\"port\":5432}",
			resourceAttrName:             "humanitec_resource_definition.postgres_test",
			configUpdate: func() string {
				return testAccResourceDefinitionPostgresResource(fmt.Sprintf("postgres-test-%d", timestamp), "test-2")
			},
			resourceAttrNameUpdateValue2: "{\"host\":\"127.0.0.1\",\"instance\":\"test:test:test\",\"name\":\"test-2\",\"port\":5432}",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "GKE",
			configCreate: func() string {
				return testAccResourceDefinitionGKEResource(fmt.Sprintf("gke-test-%d", timestamp), "test-1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("gke-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"loadbalancer\":\"1.1.1.1\",\"name\":\"test-1\",\"project_id\":\"test\",\"zone\":\"europe-west3\"}",
			resourceAttrName:             "humanitec_resource_definition.gke_test",
			configUpdate: func() string {
				return testAccResourceDefinitionGKEResource(fmt.Sprintf("gke-test-%d", timestamp), "test-2")
			},
			resourceAttrNameUpdateValue2: "{\"loadbalancer\":\"1.1.1.1\",\"name\":\"test-2\",\"project_id\":\"test\",\"zone\":\"europe-west3\"}",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "DNS",
			configCreate: func() string {
				return testAccResourceDefinitionDNSStaticResource(fmt.Sprintf("dns-test-%d", timestamp), "test-1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("dns-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"host\":\"test-1\"}",
			resourceAttrName:             "humanitec_resource_definition.dns_test",
			configUpdate: func() string {
				return testAccResourceDefinitionDNSStaticResource(fmt.Sprintf("dns-test-%d", timestamp), "test-2")
			},
			resourceAttrNameUpdateValue2: "{\"host\":\"test-2\"}",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "Ingress",
			configCreate: func() string {
				return testAccResourceDefinitionIngressResource(fmt.Sprintf("ingress-test-%d", timestamp), "test-1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("ingress-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"labels\":{\"name\":\"test-1\"},\"no_tls\":true}",
			resourceAttrName:             "humanitec_resource_definition.ingress_test",
			configUpdate: func() string {
				return testAccResourceDefinitionIngressResource(fmt.Sprintf("ingress-test-%d", timestamp), "test-2")
			},
			resourceAttrNameUpdateValue2: "{\"labels\":{\"name\":\"test-2\"},\"no_tls\":true}",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "Provision",
			configCreate: func() string {
				return testAccResourceDefinitionProvisionResource(fmt.Sprintf("provision-test-%d", timestamp), "true")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("provision-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "provision.awspolicy.match_dependents",
			resourceAttrNameUpdateValue1: "true",
			resourceAttrName:             "humanitec_resource_definition.provision_test",
			configUpdate: func() string {
				return testAccResourceDefinitionProvisionResource(fmt.Sprintf("provision-test-%d", timestamp), "false")
			},
			resourceAttrNameUpdateValue2: "false",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "k8s-logging",
			configCreate: func() string {
				return testAccResourceDefinitionK8sLoggingResource(fmt.Sprintf("k8s-logging-test-%d", timestamp), "test-1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("k8s-logging-test-%d", timestamp),
			resourceAttrNameUpdateKey:    "name",
			resourceAttrNameUpdateValue1: "test-1",
			resourceAttrName:             "humanitec_resource_definition.k8s_logging_test",
			configUpdate: func() string {
				return testAccResourceDefinitionK8sLoggingResource(fmt.Sprintf("k8s-logging-test-%d", timestamp), "test-2")
			},
			resourceAttrNameUpdateValue2: "test-2",
			importStateVerifyIgnore:      []string{"driver_inputs.secrets_string", "force_delete"},
		},
		{
			name: "S3 static - secret refs",
			configCreate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefs(fmt.Sprintf("s3-test-with-secrets-%d", timestamp), "accessKeyIdPath1", "secretAccessKeyPath1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("s3-test-with-secrets-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.secret_refs",
			resourceAttrNameUpdateValue1: "{\"aws_access_key_id\":{\"ref\":\"accessKeyIdPath1\",\"store\":\"external-secret-store\",\"version\":\"1\"},\"aws_secret_access_key\":{\"ref\":\"secretAccessKeyPath1\",\"store\":\"external-secret-store\",\"version\":\"1\"}}",
			resourceAttrName:             "humanitec_resource_definition.s3_test_with_secrets",
			configUpdate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefs(fmt.Sprintf("s3-test-with-secrets-%d", timestamp), "accessKeyIdPath2", "secretAccessKeyPath2")
			},
			resourceAttrNameUpdateValue2: "{\"aws_access_key_id\":{\"ref\":\"accessKeyIdPath2\",\"store\":\"external-secret-store\",\"version\":\"1\"},\"aws_secret_access_key\":{\"ref\":\"secretAccessKeyPath2\",\"store\":\"external-secret-store\",\"version\":\"1\"}}",
			importStateVerifyIgnore:      []string{"force_delete"},
		},
		{
			name: "S3 static - secret ref set values",
			configCreate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefValues(fmt.Sprintf("s3-test-with-secrets-%d", timestamp), "accessKeyId1", "secretAccessKey1")
			},
			resourceAttrNameIDValue:      fmt.Sprintf("s3-test-with-secrets-%d", timestamp),
			resourceAttrNameUpdateKey:    "driver_inputs.secret_refs",
			resourceAttrNameUpdateValue1: "{\"aws_access_key_id\":{\"value\":\"accessKeyId1\"},\"aws_secret_access_key\":{\"value\":\"secretAccessKey1\"}}",
			resourceAttrName:             "humanitec_resource_definition.s3_test_with_secrets",
			configUpdate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefValues(fmt.Sprintf("s3-test-with-secrets-%d", timestamp), "accessKeyId2", "secretAccessKey2")
			},
			resourceAttrNameUpdateValue2: "{\"aws_access_key_id\":{\"value\":\"accessKeyId2\"},\"aws_secret_access_key\":{\"value\":\"secretAccessKey2\"}}",
			importStateVerifyIgnore:      []string{"driver_inputs.secret_refs", "force_delete"},
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
						ImportStateVerifyIgnore: tc.importStateVerifyIgnore,
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

func TestAccResourceDefinition_S3_static_secrets(t *testing.T) {
	var expectedSecretRef string
	var expectedSecretRefAfterUpdate string
	id := fmt.Sprintf("s3-test-with-secrets-%d", time.Now().UnixNano())
	t.Run("S3 static - secrets", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Create and Read testing
				{
					Config: testAccResourceDefinitionS3taticResourceWithSecrets(id, "accessKeyId1", "secretAccessKey1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							secretRefsRaw := s.Modules[0].Resources["humanitec_resource_definition.s3_test_with_secrets"].Primary.Attributes["driver_inputs.secret_refs"]
							var secretRefs struct {
								AWSAccessKey struct {
									Ref     string `json:"ref"`
									Store   string `json:"store"`
									Version string `json:"version"`
								} `json:"aws_access_key_id"`
								AWSSecretKey struct {
									Ref     string `json:"ref"`
									Store   string `json:"store"`
									Version string `json:"version"`
								} `json:"aws_secret_access_key"`
							}

							err := json.Unmarshal([]byte(secretRefsRaw), &secretRefs)
							if err != nil {
								return err
							}

							currentVersion, err := strconv.Atoi(secretRefs.AWSAccessKey.Version)
							if err != nil {
								return err
							}

							expectedSecretRef = getDefinitionSecretRef(id, currentVersion)
							expectedSecretRefAfterUpdate = getDefinitionSecretRef(id, currentVersion+1)
							return nil
						},
						resource.TestCheckResourceAttr("humanitec_resource_definition.s3_test_with_secrets", "id", id),
						resource.TestCheckResourceAttrPtr("humanitec_resource_definition.s3_test_with_secrets", "driver_inputs.secret_refs", &expectedSecretRef),
					),
				},
				// ImportState testing
				{
					ResourceName:            "humanitec_resource_definition.s3_test_with_secrets",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"driver_inputs.secrets_string", "force_delete"},
				},
				// Update and Read testing
				{
					Config: testAccResourceDefinitionS3taticResourceWithSecrets(id, "accessKeyId2", "secretAccessKey2"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("humanitec_resource_definition.s3_test_with_secrets", "driver_inputs.secret_refs", &expectedSecretRefAfterUpdate),
					),
				},
				// Delete testing automatically occurs in TestCase
			},
		})
	})
}

func testAccResourceDefinitionS3Resource(id, region string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test" {
  id          = "%s"
  name        = "s3-test"
  type        = "s3"
  driver_type = "humanitec/s3"

  driver_inputs = {
    values_string = jsonencode({
      "region" = "%s"
    })
  }
}
`, id, region)
}

func testAccResourceDefinitionPostgresResource(id, name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "postgres_test" {
  id          = "%s"
  name        = "postgres-test"
  type        = "postgres"
  driver_type = "humanitec/postgres-cloudsql-static"

  driver_inputs = {
    values_string = jsonencode({
      "instance" = "test:test:test"
      "name" = "%s"
      "host" = "127.0.0.1"
      "port" = 5432
    })
    secrets_string = jsonencode({
      "username" = "test"
      "password" = "test"
    })
  }
}
`, id, name)
}

func testAccResourceDefinitionGKEResource(id, name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "gke_test" {
  id          = "%s"
  name        = "gke-test"
  type        = "k8s-cluster"
  driver_type = "humanitec/k8s-cluster-gke"

  driver_inputs = {
    values_string = jsonencode({
      "loadbalancer" = "1.1.1.1"
      "name" = "%s"
      "project_id" = "test"
      "zone" = "europe-west3"
		})
    secrets_string = jsonencode({
      "credentials" = {}
    })
  }
}
`, id, name)
}

func testAccResourceDefinitionDNSStaticResource(id, host string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "dns_test" {
  id          = "%s"
  name        = "dns-test"
  type        = "dns"
  driver_type = "humanitec/static"

  driver_inputs = {
    values_string = jsonencode({
      host = "%s"
    })
  }
}
`, id, host)
}

func testAccResourceDefinitionIngressResource(id, host string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "ingress_test" {
  id          = "%s"
  name        = "ingress-test"
  type        = "ingress"
  driver_type = "humanitec/ingress"

  driver_inputs = {
    values_string = jsonencode({
      labels = {
				name = "%s"
			}
			no_tls      = true
    })
  }
}
`, id, host)
}

func testAccResourceDefinitionK8sLoggingResource(id, name string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "k8s_logging_test" {
  id          = "%s"
  name        = "%s"
  type        = "logging"
  driver_type = "humanitec/logging-k8s"
}
`, id, name)
}

func testAccResourceDefinitionProvisionResource(id, matchDependents string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "provision_test" {
	id          = "%s"
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
		values_string = jsonencode({
			"region" = "us-east-1"
		})
	}
}
`, id, matchDependents)
}

func testAccResourceDefinitionS3taticResourceWithSecretRefs(id, awsAccessKeyIDPath, awsSecretAccessKeyPath string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test_with_secrets" {
  id          = "%s"
  name        = "s3-test-with-secrets"
  type        = "s3"
  driver_type = "humanitec/static"

  driver_inputs = {
	values_string = jsonencode({
      "bucket" = "test-bucket"
	  "region" = "us-east-1"
    })
    secret_refs = jsonencode({
      "aws_access_key_id"     =  {
        "ref"     = "%s"
		"store"   = "external-secret-store"
		"version" = "1"
	  }
      "aws_secret_access_key" = {
        "ref"     = "%s"
		"store"   = "external-secret-store"
		"version" = "1"
	  }
    })
  }
}
`, id, awsAccessKeyIDPath, awsSecretAccessKeyPath)
}

func testAccResourceDefinitionS3taticResourceWithSecretRefValues(id, awsAccessKeyIDValue, awsSecretAccessKeyValue string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test_with_secrets" {
  id          = "%s"
  name        = "s3-test-with-secrets"
  type        = "s3"
  driver_type = "humanitec/static"

  driver_inputs = {
	values_string = jsonencode({
      "bucket" = "test-bucket"
	  "region" = "us-east-1"
    })
    secret_refs = jsonencode({
      "aws_access_key_id"     =  {
        "value"     = "%s"
	  }
      "aws_secret_access_key" = {
        "value"     = "%s"
	  }
    })
  }
}
`, id, awsAccessKeyIDValue, awsSecretAccessKeyValue)
}

func testAccResourceDefinitionS3taticResourceWithSecrets(id string, awsAccessKeyIDValue, awsSecretAccessKeyValue string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test_with_secrets" {
  id          = "%s"
  name        = "s3-test-with-secrets"
  type        = "s3"
  driver_type = "humanitec/static"

  driver_inputs = {
	values_string = jsonencode({
      "bucket" = "test-bucket"
	  "region" = "us-east-1"
    })
    secrets_string = jsonencode({
      "aws_access_key_id"     = "%s"
      "aws_secret_access_key" = "%s"
    })
  }
}
`, id, awsAccessKeyIDValue, awsSecretAccessKeyValue)
}

func getDefinitionSecretPath(defID string) string {
	orgID := os.Getenv("HUMANITEC_ORG")
	return fmt.Sprintf("orgs/%s/resources/defs/%s/driver_secrets", orgID, defID)
}

func getDefinitionSecretRef(id string, version int) string {
	return fmt.Sprintf("{\"aws_access_key_id\":{\"ref\":\"%s/aws_access_key_id/.value\",\"store\":\"humanitec\",\"version\":\"%d\"},\"aws_secret_access_key\":{\"ref\":\"%s/aws_secret_access_key/.value\",\"store\":\"humanitec\",\"version\":\"%d\"}}", getDefinitionSecretPath(id), version, getDefinitionSecretPath(id), version)
}
