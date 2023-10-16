resource "humanitec_secretstore" "secret_store_gcpsm" {
  id = "secretstore_id"
  gcpsm = {
    project_id = "example-project"
    auth = {
      secret_access_key = "secret-access-key"
    }
  }
}
