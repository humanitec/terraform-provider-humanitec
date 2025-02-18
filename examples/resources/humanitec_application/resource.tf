resource "humanitec_application" "example" {
  id   = "example"
  name = "An example app"
}

resource "humanitec_environment" "example" {
  app_id = humanitec_application.example.id
  id     = "development"
  name   = "Development"
  type   = "development"
}
