resource "humanitec_resource_definition" "example" {
  id   = "example-s3"
  name = "example-s3"
  type = "s3"

  driver_type = "humanitec/s3"
  driver_inputs = {
    values = {
      region = "us-east-1"
    }
  }

  lifecycle {
    ignore_changes = [
      criteria
    ]
  }
}

resource "humanitec_resource_definition_criteria" "example" {
  resource_definition_id = humanitec_resource_definition.example.id
  app_id                 = "example-app"
}
