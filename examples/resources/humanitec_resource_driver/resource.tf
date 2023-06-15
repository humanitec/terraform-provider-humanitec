resource "humanitec_resource_driver" "s3" {
  id   = "demo-driver"
  type = "s3"

  account_types = [
    "aws",
  ]

  inputs_schema = jsonencode({})
  target        = "https://drivers.example.com/s3/"
}
