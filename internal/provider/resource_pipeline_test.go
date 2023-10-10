package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourcePipeline(t *testing.T) {
	appID := fmt.Sprintf("test-%d", time.Now().UnixNano())
	definition := `
name: Hello from terraform
on: 
  pipeline_call:
jobs:
  approve:
    steps:
    - name: approve
      uses: actions/humanitec/approve
      with:
        environment: development
        message: Test message
`
	newDefinition := `
name: Hello from terraform - update
on: 
  pipeline_call:
jobs:
  approve:
    steps:
    - name: approve
      uses: actions/humanitec/approve
      with:
        environment: development
        message: Test message
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourcePipeline(appID, definition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_pipeline.pipeline_test", "app_id", appID),
					resource.TestCheckResourceAttr("humanitec_pipeline.pipeline_test", "definition", definition+"\n"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "humanitec_application.app_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "humanitec_pipeline.pipeline_test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					pipeline, err := testResource("humanitec_pipeline.pipeline_test", s)
					if err != nil {
						return "", err
					}

					return fmt.Sprintf("%s/%s", appID, pipeline.Primary.ID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccResourcePipeline(appID, newDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("humanitec_pipeline.pipeline_test", "app_id", appID),
					resource.TestCheckResourceAttr("humanitec_pipeline.pipeline_test", "definition", newDefinition+"\n"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccResourcePipeline(app, definition string) string {
	return fmt.Sprintf(`
resource "humanitec_application" "app_test" {
	id          = "%s"
	name        = "test-app"
}

resource "humanitec_pipeline" "pipeline_test" {
	app_id     = humanitec_application.app_test.id
	definition = <<EOT
%s
EOT
}`, app, definition)
}
