package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourcePipelineCriteria(t *testing.T) {
	// avoid conflict by giving apps a unique id
	testUid := int(time.Now().UnixMilli())

	// base config contains the "static" bits for this test that don't change
	baseConfig := fmt.Sprintf(`
resource humanitec_application "app" {
	id = "app%[1]d"
	name = "App %[1]d"
}

resource humanitec_pipeline "pip" {
	app_id = humanitec_application.app.id
	definition = <<EOT
name: Test pipeline
on:
  deployment_request: {}
jobs:
  thing:
    steps:
    - uses: actions/humanitec/log
      with:
        message: $${{ tojson(inputs) }}
EOT
}
`, testUid)
	criteriaV1 := `
resource humanitec_pipeline_criteria "c1" {
	app_id = humanitec_application.app.id
	pipeline_id = humanitec_pipeline.pip.id
    deployment_request = {
        env_id = "development"
        deployment_type = "deploy"
    }
}
`
	criteriaV2 := `
resource humanitec_pipeline_criteria "c1" {
	app_id = humanitec_application.app.id
	pipeline_id = humanitec_pipeline.pip.id
    deployment_request = {
		env_type = "development"
        deployment_type = "re-deploy"
    }
}
`

	resource.Test(t, resource.TestCase{
		// check whether env vars are set
		PreCheck: func() { testAccPreCheck(t) },
		// get the humanitec provider for tests
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// first test that we can create a app, pipeline, and pipeline criteria via the TF resource.
			{
				Config: baseConfig + criteriaV1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "id"),
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "pipeline_id"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "pipeline_name", "Test pipeline"),
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "deployment_request.app_id"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "deployment_request.env_id", "development"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "deployment_request.deployment_type", "deploy"),
				),
			},
			// test that another plan should be empty
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "id"),
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "pipeline_id"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "pipeline_name", "Test pipeline"),
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "deployment_request.app_id"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "deployment_request.env_id", "development"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "deployment_request.deployment_type", "deploy"),
				),
			},
			// now we can test that we can edit the criteria and that it should be deleted and recreated
			{
				Config: baseConfig + criteriaV2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "id"),
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "pipeline_id"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "pipeline_name", "Test pipeline"),
					resource.TestCheckResourceAttrSet("humanitec_pipeline_criteria.c1", "deployment_request.app_id"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "deployment_request.env_type", "development"),
					resource.TestCheckResourceAttr("humanitec_pipeline_criteria.c1", "deployment_request.deployment_type", "re-deploy"),
				),
			},
			// now let's test that we can import things reliably
			{
				ResourceName: "humanitec_pipeline_criteria.c1",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					for _, module := range state.Modules {
						res, ok := module.Resources["humanitec_pipeline_criteria.c1"]
						if ok {
							return fmt.Sprintf("%s/%s/%s", res.Primary.Attributes["app_id"], res.Primary.Attributes["pipeline_id"], res.Primary.ID), nil
						}
					}
					return "", fmt.Errorf("failed to find resource in state")
				},
				ImportStateVerify: true,
			},
		},
	})
}
