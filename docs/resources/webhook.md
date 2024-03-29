---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "humanitec_webhook Resource - terraform-provider-humanitec"
subcategory: ""
description: |-
  Webhook is a special type of a Job, it performs a HTTPS request to a specified URL with specified headers.
---

# humanitec_webhook (Resource)

Webhook is a special type of a Job, it performs a HTTPS request to a specified URL with specified headers.

## Example Usage

```terraform
resource "humanitec_webhook" "webhook1" {
  id     = "my-hook"
  app_id = "app-id"

  url = "https://example.com/hook"
  triggers = [{
    scope = "environment"
    type  = "created"
  }]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_id` (String) The ID of the Application that the Webhook should belong to.
- `id` (String) The ID of the Webhook.
- `triggers` (Attributes Set) A list of Events by which the Job is triggered, supported triggers are:

  | scope | type |
	|-------|------|
	| environment  | created |
	| environment  | deleted |
	| deployment  | started |
	| deployment  | finished | (see [below for nested schema](#nestedatt--triggers))
- `url` (String) Thw webhook's URL (without protocol, only HTTPS is supported)

### Optional

- `disabled` (Boolean) Defines whether this job is currently disabled.
- `headers` (Map of String) Custom webhook headers.
- `payload` (Map of String) Customize payload.

<a id="nestedatt--triggers"></a>
### Nested Schema for `triggers`

Required:

- `scope` (String) Scope of the trigger
- `type` (String) Type of the trigger

## Import

Import is supported using the following syntax:

```shell
# import an existing webhook
terraform import humanitec_webhook.my_hook app_id/webhook_id
```
