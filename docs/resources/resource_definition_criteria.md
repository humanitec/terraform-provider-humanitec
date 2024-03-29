---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "humanitec_resource_definition_criteria Resource - terraform-provider-humanitec"
subcategory: ""
description: |-
  Visit the docs https://docs.humanitec.com/reference/concepts/resources/definitions to learn more about resource definitions.
---

# humanitec_resource_definition_criteria (Resource)

Visit the [docs](https://docs.humanitec.com/reference/concepts/resources/definitions) to learn more about resource definitions.

## Example Usage

```terraform
resource "humanitec_resource_definition" "example" {
  id   = "example-s3"
  name = "example-s3"
  type = "s3"

  driver_type = "humanitec/s3"
  driver_inputs = {
    values_string = jsonencode({
      region = "us-east-1"
    })
  }

  lifecycle {
    ignore_changes = [
      criteria
    ]
  }
}

resource "humanitec_resource_definition_criteria" "example" {
  resource_definition_id = humanitec_resource_definition.example.id
  app_id                 = "example-app"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `resource_definition_id` (String) The Resource Definition ID.

### Optional

- `app_id` (String) The ID of the Application that the Resources should belong to.
- `class` (String) The class of the Resource in the Deployment Set. Can not be empty, if is not defined, set to `default`.
- `env_id` (String) The ID of the Environment that the Resources should belong to. If `env_type` is also set, it must match the Type of the Environment for the Criteria to match.
- `env_type` (String) The Type of the Environment that the Resources should belong to. If `env_id` is also set, it must have an Environment Type that matches this parameter for the Criteria to match.
- `force_delete` (Boolean) If set to `true`, the Matching Criteria is deleted immediately, even if this action affects existing Active Resources.
- `res_id` (String) The ID of the Resource in the Deployment Set. The ID is normally a `.` separated path to the definition in the set, e.g. `modules.my-module.externals.my-database`.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Matching Criteria ID

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.

## Import

Import is supported using the following syntax:

```shell
terraform import humanitec_resource_definition_criteria.example resource_definition_id/criteria_id
```
