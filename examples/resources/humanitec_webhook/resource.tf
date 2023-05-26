resource "humanitec_webhook" "webhook1" {
  id     = "my-hook"
  app_id = "app-id"

  url = "https://example.com/hook"
  triggers = [{
    scope = "environment"
    type  = "created"
  }]
}
