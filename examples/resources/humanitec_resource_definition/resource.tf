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

resource "humanitec_resource_definition" "dns" {
  id   = "dns-newapp"
  name = "dns-newapp"
  type = "dns"

  driver_type = "humanitec/newapp-io-dns"

  provision = {
    "ingress" = {
      is_dependent     = false
      match_dependents = false
      params = jsonencode({
        "host" = "$${resources.dns.host}"
      })
    }
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

resource "humanitec_resource_definition" "azure-blob" {
  driver_type = "humanitec/terraform"
  id          = "azure-blob"
  name        = "azure-blob"
  type        = "azure-blob"

  driver_inputs = {
    secret_refs = jsonencode({
      variables = {
        client_id = {
          ref   = var.client_id_secret_reference_key
          store = var.secret_store
        }
        client_secret = {
          ref   = var.client_secret_secret_reference_key
          store = var.secret_store
        }
      }

      source = {
        ssh_key = {
          ref   = var.ssh_key_secret_reference_key
          store = var.secret_store
        }
      }
    })

    values_string = jsonencode({
      source = {
        path = var.tf_module_github_path
        rev  = var.tf_module_github_ref
        url  = var.tf_module_github_url
      }

      variables = {
        tenant_id           = var.tenant_id
        subscription_id     = var.subscription_id
        resource_group_name = var.resource_group_name
      }
    })
  }
}
