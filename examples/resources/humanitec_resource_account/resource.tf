resource "humanitec_resource_account" "gcp_test" {
  id          = "gcp-dev"
  name        = "gcp-dev"
  type        = "gcp"
  credentials = "{ json from key }"
}
