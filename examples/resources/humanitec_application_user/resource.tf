resource "humanitec_application_user" "another_owner" {
  app_id  = "example"
  user_id = "user-id"
  role    = "owner"
}
