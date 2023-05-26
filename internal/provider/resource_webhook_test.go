package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceWebhook(t *testing.T) {

	testCases := []struct {
		name   string
		config func(appId, url string) string
	}{
		{
			name: "basic",
			config: func(appId, url string) string {
				return testAccResourceWebhook(appId, url)
			},
		},
		{
			name: "full",
			config: func(appId, url string) string {
				return testAccResourceWebhook_Full(appId, url)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appId := fmt.Sprintf("tf-webhook-%d", time.Now().UnixNano())

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read testing
					{
						Config: tc.config(appId, "https://example.com"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_webhook.webhook1", "id", "my-hook"),
							resource.TestCheckResourceAttr("humanitec_webhook.webhook1", "url", "https://example.com"),
						),
					},
					// ImportState testing
					{
						ResourceName: "humanitec_webhook.webhook1",
						ImportState:  true,
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							return fmt.Sprintf("%s/%s", appId, "my-hook"), nil
						},
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{},
					},
					// Update and Read testing
					{
						Config: tc.config(appId, "https://example2.com"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("humanitec_webhook.webhook1", "id", "my-hook"),
							resource.TestCheckResourceAttr("humanitec_webhook.webhook1", "url", "https://example2.com"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

func testAccResourceWebhook(id, url string) string {
	return fmt.Sprintf(`
	resource "humanitec_application" "webhook_test" {
		id   = "%s"
		name = "webhook-test"
	}

	resource "humanitec_webhook" "webhook1" {
		id     = "my-hook"
		app_id = humanitec_application.webhook_test.id

		url =  "%s"
		triggers = [{
			scope = "environment"
			type = "created"
		}]
	}
`, id, url)
}

func testAccResourceWebhook_Full(id, url string) string {
	return fmt.Sprintf(`
	resource "humanitec_application" "webhook_test" {
		id   = "%s"
		name = "webhook-test"
	}

	resource "humanitec_webhook" "webhook1" {
		id     = "my-hook"
		app_id = humanitec_application.webhook_test.id

		url =  "%s"
		triggers = [{
			scope = "environment"
			type = "created"
		}]

		headers = {
			"custom-header" = "humanitec"
		}

		payload = {
			"custom-payload" = "humanitec"
		}

		disabled = true
	}
`, id, url)
}
