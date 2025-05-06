data "humanitec_user_groups" "admins" {
  filter = {
    idp_id   = "sample-idp"
    group_id = "sample-group"
  }
}

data "humanitec_user_groups" "members" {
  filter = {
    id = "g-1234567890"
  }
}
