resource "humanitec_resource_definition" "s3" {
  id   = "test-s3"
  name = "test-s3"
  type = "s3"

  driver_type = "humanitec/s3"
  driver_inputs = {
    values = {
      region = "us-east-1"
    }
  }
}
