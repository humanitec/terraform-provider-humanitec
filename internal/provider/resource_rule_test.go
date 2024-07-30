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

func TestAccResourceRule_DeletedManually(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	appId := fmt.Sprintf("tf-rule-%d", time.Now().UnixNano())

	orgID := os.Getenv("HUMANITEC_ORG")
	token := os.Getenv("HUMANITEC_TOKEN")
	apiHost := os.Getenv("HUMANITEC_HOST")
	if apiHost == "" {
		apiHost = humanitec.DefaultAPIHost
	}

	var client *humanitec.Client
	var err error
	var id string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			client, err = NewHumanitecClient(apiHost, token, "test", nil)
			assert.NoError(err)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRule(appId, "my-artefact"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_rule.rule1", "artefacts_filter.0", "my-artefact"),
				),
			},
			{
				ResourceName:      "humanitec_rule.rule1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rule, err := testResource("humanitec_rule.rule1", s)
					if err != nil {
						return "", err
					}
					id = fmt.Sprintf("%s/dev/%s", appId, rule.Primary.ID)
					return id, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_rule.rule1", "id", id),
					func(_ *terraform.State) error {
						// Manually delete the rule via the API
						resp, err := client.DeleteAutomationRuleWithResponse(ctx, orgID, appId, "dev", id)
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
