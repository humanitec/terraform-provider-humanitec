resource "humanitec_user_group" "example-group" {
  group_id = "administrators"
  role     = "administrator"
  idp_id   = "sample-idp-id"
}
