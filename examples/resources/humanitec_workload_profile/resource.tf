resource "humanitec_workload_profile" "custom" {
  id          = "custom-profile"
  description = "Custom workload profile"
  version     = "1.0.0"

  # See https://developer.humanitec.com/integration-and-extensions/workload-profiles/custom-workload-profiles/
  # for more information on the spec_definition
  spec_definition = jsonencode({})

  workload_profile_chart = {
    id      = "humanitec/default-module"
    version = "1.0.0"
  }
}
