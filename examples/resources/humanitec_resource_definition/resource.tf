resource "humanitec_resource_definition" "s3" {
  id   = "s3-dev-bucket"
  name = "s3-dev-bucket"
  type = "s3"

  driver_type = "humanitec/s3"
  driver_inputs = {
    values = {
      region = "us-east-1"
    }
  }
}

resource "humanitec_resource_definition" "postgres" {
  id          = "db-dev"
  name        = "db-dev"
  type        = "postgres"
  driver_type = "humanitec/postgres-cloudsql-static"

  driver_inputs = {
    values = {
      "instance" = "test:test:test"
      "name"     = "db-dev"
      "host"     = "127.0.0.1"
      "port"     = "5432"
    }
    secrets = {
      "username" = "test"
      "password" = "test"
    }
  }

  criteria = [{
    app_id = "test-app"
  }]
}

resource "humanitec_resource_definition" "gke" {
  id          = "gke-dev"
  name        = "gke-dev"
  type        = "k8s-cluster"
  driver_type = "humanitec/k8s-cluster-gke"

  driver_inputs = {
    values = {
      "loadbalancer" = "1.1.1.1"
      "name"         = "gke-dev"
      "project_id"   = "test"
      "zone"         = "europe-west3"
    }
    secrets = {
      "credentials" = "{}"
    }
  }

  criteria = [
    {
      app_id   = "test-app"
      env_type = "development"
    },
    {
      app_id   = "test-app"
      env_type = "staging"
    }
  ]
}
