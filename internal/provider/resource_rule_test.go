package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceRule(t *testing.T) {

	testCases := []struct {
		name   string
		config func(appId, url string) string
	}{
		{
			name: "basic",
			config: func(appId, url string) string {
				return testAccResourceRule(appId, url)
			},
		},
		{
			name: "full",
			config: func(appId, url string) string {
				return testAccResourceRule_Full(appId, url)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appId := fmt.Sprintf("tf-rule-%d", time.Now().UnixNano())

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: tc.config(appId, "my-artefact"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_rule.rule1", "artefacts_filter.0", "my-artefact"),
						),
					},
					// ImportState testing
					{
						ResourceName: "humanitec_rule.rule1",
						ImportState:  true,
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							rule, err := testResource("humanitec_rule.rule1", s)
							if err != nil {
								return "", err
							}

							return fmt.Sprintf("%s/dev/%s", appId, rule.Primary.ID), nil
						},
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{},
					},
					// Update and Read testing
					{
						Config: tc.config(appId, "my-artefact2"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_rule.rule1", "artefacts_filter.0", "my-artefact2"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceRule(appID, artefact string) string {
	return fmt.Sprintf(`
	resource "humanitec_application" "rule_test" {
		id = "%s"
		name = "rule-test"

		env = {
			id   = "dev"
			name = "dev"
			type = "development"
		}
	}

	resource "humanitec_rule" "rule1" {
		app_id = humanitec_application.rule_test.id
		env_id = "dev"

		artefacts_filter = ["%s"]
		match_ref = "refs/main"
		type = "update"
	}
`, appID, artefact)
}

func testAccResourceRule_Full(appID, artefact string) string {
	return fmt.Sprintf(`
	resource "humanitec_application" "rule_test" {
		id = "%s"
		name = "rule-test"

		env = {
			id   = "dev"
			name = "dev"
			type = "development"
		}
	}

	resource "humanitec_rule" "rule1" {
		app_id = humanitec_application.rule_test.id
		env_id = "dev"

		active = false
		artefacts_filter = ["%s"]
		exclude_artefacts_filter = true
		match_ref = "refs/main"
		type = "update"
	}
`, appID, artefact)
}
