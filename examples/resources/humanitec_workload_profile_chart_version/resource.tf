resource "humanitec_workload_profile_chart_version" "custom" {
  filename         = "./chart.tar.gz"
  source_code_hash = filebase64sha256("./chart.tar.gz")
}

# The Chart Version can be referenced in the Workload Profile
resource "humanitec_workload_profile" "custom" {
  id          = "custom-profile"
  description = "Custom workload profile"
  version     = "1.0.0"

  # See https://developer.humanitec.com/integration-and-extensions/workload-profiles/custom-workload-profiles/
  # for more information on the spec_definition
  spec_definition = jsonencode({})

  workload_profile_chart = {
    id      = humanitec_workload_profile_chart_version.custom.id
    version = humanitec_workload_profile_chart_version.custom.version
  }
}
