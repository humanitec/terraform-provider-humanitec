package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"region\":\"us-east-1\"}",
			resourceAttrName:             "humanitec_resource_definition.s3_test",
			configUpdate: func() string {
				return testAccResourceDefinitionS3Resource("us-east-2")
			},
			resourceAttrNameUpdateValue2: "{\"region\":\"us-east-2\"}",
		},
		{
			name: "Postgres",
			configCreate: func() string {
				return testAccResourceDefinitionPostgresResource("test-1")
			},
			resourceAttrNameIDValue:      "postgres-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"host\":\"127.0.0.1\",\"instance\":\"test:test:test\",\"name\":\"test-1\",\"port\":5432}",
			resourceAttrName:             "humanitec_resource_definition.postgres_test",
			configUpdate: func() string {
				return testAccResourceDefinitionPostgresResource("test-2")
			},
			resourceAttrNameUpdateValue2: "{\"host\":\"127.0.0.1\",\"instance\":\"test:test:test\",\"name\":\"test-2\",\"port\":5432}",
		},
		{
			name: "GKE",
			configCreate: func() string {
				return testAccResourceDefinitionGKEResource("test-1")
			},
			resourceAttrNameIDValue:      "gke-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"loadbalancer\":\"1.1.1.1\",\"name\":\"test-1\",\"project_id\":\"test\",\"zone\":\"europe-west3\"}",
			resourceAttrName:             "humanitec_resource_definition.gke_test",
			configUpdate: func() string {
				return testAccResourceDefinitionGKEResource("test-2")
			},
			resourceAttrNameUpdateValue2: "{\"loadbalancer\":\"1.1.1.1\",\"name\":\"test-2\",\"project_id\":\"test\",\"zone\":\"europe-west3\"}",
		},
		{
			name: "DNS",
			configCreate: func() string {
				return testAccResourceDefinitionDNSStaticResource("test-1")
			},
			resourceAttrNameIDValue:      "dns-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"host\":\"test-1\"}",
			resourceAttrName:             "humanitec_resource_definition.dns_test",
			configUpdate: func() string {
				return testAccResourceDefinitionDNSStaticResource("test-2")
			},
			resourceAttrNameUpdateValue2: "{\"host\":\"test-2\"}",
		},
		{
			name: "Ingress",
			configCreate: func() string {
				return testAccResourceDefinitionIngressResource("test-1")
			},
			resourceAttrNameIDValue:      "ingress-test",
			resourceAttrNameUpdateKey:    "driver_inputs.values_string",
			resourceAttrNameUpdateValue1: "{\"labels\":{\"name\":\"test-1\"},\"no_tls\":true}",
			resourceAttrName:             "humanitec_resource_definition.ingress_test",
			configUpdate: func() string {
				return testAccResourceDefinitionIngressResource("test-2")
			},
			resourceAttrNameUpdateValue2: "{\"labels\":{\"name\":\"test-2\"},\"no_tls\":true}",
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
		{
			name: "S3 static - secret refs",
			configCreate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefs("accessKeyIdPath1", "secretAccessKeyPath1")
			},
			resourceAttrNameIDValue:      "s3-test-with-secrets",
			resourceAttrNameUpdateKey:    "driver_inputs.secret_refs",
			resourceAttrNameUpdateValue1: "{\"aws_access_key_id\":{\"ref\":\"accessKeyIdPath1\",\"store\":\"external-secret-store\",\"version\":\"1\"},\"aws_secret_access_key\":{\"ref\":\"secretAccessKeyPath1\",\"store\":\"external-secret-store\",\"version\":\"1\"}}",
			resourceAttrName:             "humanitec_resource_definition.s3_test_with_secrets",
			configUpdate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefs("accessKeyIdPath2", "secretAccessKeyPath2")
			},
			resourceAttrNameUpdateValue2: "{\"aws_access_key_id\":{\"ref\":\"accessKeyIdPath2\",\"store\":\"external-secret-store\",\"version\":\"1\"},\"aws_secret_access_key\":{\"ref\":\"secretAccessKeyPath2\",\"store\":\"external-secret-store\",\"version\":\"1\"}}",
		},
		{
			name: "S3 static - secret ref set values",
			configCreate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefValues("accessKeyId1", "secretAccessKey1")
			},
			resourceAttrNameIDValue:      "s3-test-with-secrets",
			resourceAttrNameUpdateKey:    "driver_inputs.secret_refs",
			resourceAttrNameUpdateValue1: "{\"aws_access_key_id\":{\"value\":\"accessKeyId1\"},\"aws_secret_access_key\":{\"value\":\"secretAccessKey1\"}}",
			resourceAttrName:             "humanitec_resource_definition.s3_test_with_secrets",
			configUpdate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecretRefValues("accessKeyId2", "secretAccessKey2")
			},
			resourceAttrNameUpdateValue2: "{\"aws_access_key_id\":{\"value\":\"accessKeyId2\"},\"aws_secret_access_key\":{\"value\":\"secretAccessKey2\"}}",
		},
		{
			name: "S3 static - secrets",
			configCreate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecrets("accessKeyId1", "secretAccessKey1")
			},
			resourceAttrNameIDValue:      "s3-test-with-secrets",
			resourceAttrNameUpdateKey:    "driver_inputs.secret_refs",
			resourceAttrNameUpdateValue1: fmt.Sprintf("{\"aws_access_key_id\":{\"ref\":\"%s/aws_access_key_id/.value\",\"store\":\"humanitec\"},\"aws_secret_access_key\":{\"ref\":\"%s/aws_secret_access_key/.value\",\"store\":\"humanitec\"}}", getDefinitionSecretPath("s3-test-with-secrets"), getDefinitionSecretPath("s3-test-with-secrets")),
			resourceAttrName:             "humanitec_resource_definition.s3_test_with_secrets",
			configUpdate: func() string {
				return testAccResourceDefinitionS3taticResourceWithSecrets("accessKeyId2", "secretAccessKey2")
			},
			resourceAttrNameUpdateValue2: fmt.Sprintf("{\"aws_access_key_id\":{\"ref\":\"%s/aws_access_key_id/.value\",\"store\":\"humanitec\"},\"aws_secret_access_key\":{\"ref\":\"%s/aws_secret_access_key/.value\",\"store\":\"humanitec\"}}", getDefinitionSecretPath("s3-test-with-secrets"), getDefinitionSecretPath("s3-test-with-secrets")),
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
						ImportStateVerifyIgnore: []string{"driver_inputs.secrets_string", "driver_inputs.secret_refs", "force_delete"},
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
    values_string = jsonencode({
      "region" = "%s"
    })
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
    values_string = jsonencode({
      host = "%s"
    })
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
    values_string = jsonencode({
      labels = {
				name = "%s"
			}
			no_tls      = true
    })
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
		values_string = jsonencode({
			"region" = "us-east-1"
		})
	}
}
`, matchDependents)
}

func testAccResourceDefinitionS3taticResourceWithSecretRefs(awsAccessKeyIDPath, awsSecretAccessKeyPath string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test_with_secrets" {
  id          = "s3-test-with-secrets"
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
`, awsAccessKeyIDPath, awsSecretAccessKeyPath)
}

func testAccResourceDefinitionS3taticResourceWithSecretRefValues(awsAccessKeyIDValue, awsSecretAccessKeyValue string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test_with_secrets" {
  id          = "s3-test-with-secrets"
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
`, awsAccessKeyIDValue, awsSecretAccessKeyValue)
}

func testAccResourceDefinitionS3taticResourceWithSecrets(awsAccessKeyIDValue, awsSecretAccessKeyValue string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test_with_secrets" {
  id          = "s3-test-with-secrets"
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
`, awsAccessKeyIDValue, awsSecretAccessKeyValue)
}

func getDefinitionSecretPath(defID string) string {
	orgID := os.Getenv("HUMANITEC_ORG")
	return fmt.Sprintf("orgs/%s/resources/defs/%s/driver_secrets", orgID, defID)
}
