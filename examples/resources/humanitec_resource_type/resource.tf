resource "humanitec_resource_type" "demo" {
  id       = "demo-type"
  name     = "Demo Type"
  category = "Demo category"
  use      = "direct"
  inputs_schema = jsonencode({
    "additionalProperties" : false,
    "type" : "object"
  })
  outputs_schema = jsonencode({
    "properties" : {
      "values" : {
        "properties" : {
          "host" : {
            "description" : "The IP address or hostname",
            "title" : "Host",
            "type" : "string"
          },
          "port" : {
            "description" : "The port on the host",
            "maximum" : 65535,
            "minimum" : 0,
            "title" : "Port",
            "type" : "integer"
          }
        },
        "required" : ["host", "port"],
        "type" : "object"
      }
    },
    "type" : "object"
  })
}
