resource "humanitec_pipeline_criteria" "example" {
  app_id      = humanitec_application.example.id
  pipeline_id = humanitec_pipeline.example.id
  deployment_request = {
    env_type        = "development"
    deployment_type = "deploy"
  }
}
