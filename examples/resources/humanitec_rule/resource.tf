resource "humanitec_rule" "rule1" {
  id     = "my-hook"
  app_id = "app-id"
  env_id = "dev"

  artefacts_filter = ["my-org/my-image"]
  match_ref        = "refs/main"
  type             = "update"
}
