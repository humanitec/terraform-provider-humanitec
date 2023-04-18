resource "humanitec_value" "app_val1" {
  app_id = "example-app"

  key         = "VAL1"
  description = "app level value"
  value       = "EXAMPLE"
  is_secret   = false
}

resource "humanitec_value" "app_env_val1" {
  app_id = "example-app"
  env_id = "production"

  key         = "VAL1"
  description = "app env level value"
  value       = "EXAMPLE"
  is_secret   = false
}
