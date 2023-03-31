package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDefinitionCriteria(t *testing.T) {
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
			name: "WithAppID",
			configCreate: func() string {
				return testAccResourceDefinitionAndCriteriaResource("my-app-1")
			},
			resourceAttrNameIDValue:      "s3-test",
			resourceAttrNameUpdateKey:    "app_id",
			resourceAttrNameUpdateValue1: "my-app-1",
			resourceAttrName:             "humanitec_resource_definition_criteria.s3_test",
			configUpdate: func() string {
				return testAccResourceDefinitionAndCriteriaResource("my-app-2")
			},
			resourceAttrNameUpdateValue2: "my-app-2",
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
							resource.TestCheckResourceAttr(tc.resourceAttrName, "resource_definition_id", tc.resourceAttrNameIDValue),
							resource.TestCheckResourceAttr(tc.resourceAttrName, tc.resourceAttrNameUpdateKey, tc.resourceAttrNameUpdateValue1),
						),
					},
					// ImportState testing
					{
						ResourceName: tc.resourceAttrName,
						ImportState:  true,
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							criteria, err := testResource(tc.resourceAttrName, s)
							if err != nil {
								return "", err
							}

							return fmt.Sprintf("s3-test/%s", criteria.Primary.ID), nil
						},
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"force_delete"},
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

func testResource(resourceName string, state *terraform.State) (*terraform.ResourceState, error) {
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				return v, nil
			}
		}
	}

	return nil, fmt.Errorf(
		"Resource specified by ResourceName couldn't be found: %s", resourceName)
}

func testAccResourceDefinitionAndCriteriaResource(appID string) string {
	return fmt.Sprintf(`
resource "humanitec_resource_definition" "s3_test" {
  id          = "s3-test"
  name        = "s3-test"
  type        = "s3"
  driver_type = "humanitec/s3"

  driver_inputs = {
    values = {
      "region" = "us-east-1"
    }
  }

	lifecycle {
    ignore_changes = [
      criteria
    ]
  }
}

resource "humanitec_resource_definition_criteria" "s3_test" {
  resource_definition_id = humanitec_resource_definition.s3_test.id
	app_id = "%s"

}
`, appID)
}
