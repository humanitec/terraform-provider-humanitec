resource "humanitec_registry" "example" {
  id        = "example-registry"
  registry  = "registry.example.com"
  type      = "secret_ref"
  enable_ci = true
  secrets = {
    cluster-a = {
      namespace = "example-namespace"
      secret    = "path/to/secret"
    }
  }
}
