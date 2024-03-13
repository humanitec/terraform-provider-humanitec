package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceWorkloadProfile(t *testing.T) {
	testCases := []struct {
		name   string
		config func(id, description, version string) string
	}{
		{
			name:   "basic",
			config: testAccResourceWorkloadProfile,
		},
		{
			name:   "full",
			config: testAccResourceWorkloadProfile_Full,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workloadProfileID := fmt.Sprintf("profile-%d", time.Now().UnixNano())

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: tc.config(workloadProfileID, "desc1", "1.0.0"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_workload_profile.main", "id", workloadProfileID),
							resource.TestCheckResourceAttr("humanitec_workload_profile.main", "description", "desc1"),
						),
					},
					// ImportState testing
					{
						ResourceName:            "humanitec_workload_profile.main",
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{},
					},
					// Update and Read testing
					{
						Config: tc.config(workloadProfileID, "desc2", "1.0.1"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_workload_profile.main", "id", workloadProfileID),
							resource.TestCheckResourceAttr("humanitec_workload_profile.main", "description", "desc2"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceWorkloadProfile(id, description, version string) string {
	return fmt.Sprintf(`
resource "humanitec_workload_profile" "main" {
	id = "%s"
	description = "%s"
	spec_definition = jsonencode({})
	version = "%s"
	workload_profile_chart = {
		id = "humanitec/default-module"
		version = "latest"
	}
}
`, id, description, version)
}

func testAccResourceWorkloadProfile_Full(id, description, version string) string {
	return fmt.Sprintf(`
resource "humanitec_workload_profile" "main" {
	id = "%s"
	deprecation_message = "deprecation message"
	description = "%s"
	spec_definition = jsonencode({})
	version = "%s"
	workload_profile_chart = {
		id = "humanitec/default-module"
		version = "latest"
	}
}
`, id, description, version)
}
