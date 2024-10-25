resource "humanitec_application" "example" {
  id   = "example"
  name = "An example app"
}

resource "humanitec_application" "example" {
  id   = "example"
  name = "An example app with default development environment overriden"

  env = {
    id   = "dev"
    name = "Dev"
    type = "development"
  }
}