resource "humanitec_pipeline" "example" {
  app_id     = "example-app"
  definition = <<EOT
name: Example pipeline
on: 
  pipeline_call:
jobs:
  log:
    steps:
    - name: log
      uses: actions/humanitec/log
      with:
        message: Hello from Terraform
EOT
}
