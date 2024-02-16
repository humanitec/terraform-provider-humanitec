resource "humanitec_resource_definition" "s3" {
  id   = "s3-dev-bucket"
  name = "s3-dev-bucket"
  type = "s3"

  driver_type = "humanitec/s3"
  driver_inputs = {
    values_string = jsonencode({
      region = "us-east-1"
    })
  }
}

resource "humanitec_resource_definition" "postgres" {
  id          = "db-dev"
  name        = "db-dev"
  type        = "postgres"
  driver_type = "humanitec/postgres-cloudsql-static"

  driver_inputs = {
    values_string = jsonencode({
      "instance" = "test:test:test"
      "name"     = "db-dev"
      "host"     = "127.0.0.1"
      "port"     = "5432"
    })
    secrets_string = jsonencode({
      "username" = "test"
      "password" = "test"
    })
  }
}

resource "humanitec_resource_definition" "gke" {
  id          = "gke-dev"
  name        = "gke-dev"
  type        = "k8s-cluster"
  driver_type = "humanitec/k8s-cluster-gke"

  driver_inputs = {
    values_string = jsonencode({
      "loadbalancer" = "1.1.1.1"
      "name"         = "gke-dev"
      "project_id"   = "test"
      "zone"         = "europe-west3"
    })
    secrets_string = jsonencode({
      "credentials" = "{}"
    })
  }
}
