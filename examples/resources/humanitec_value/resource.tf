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

# In case you want to import a secret value, but don't want the secret in your state file, you can ignore the value field.
resource "humanitec_value" "app_val1_ignore_value" {
  app_id = "example-app"

  key         = "VAL1"
  description = "app env level value"
  is_secret   = true

  lifecycle {
    ignore_changes = [
      value,
    ]
  }
}
