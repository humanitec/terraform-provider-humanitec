resource "humanitec_environment_type_user" "another_deployer" {
  env_type_id = "production"
  user_id     = "user-id"
  role        = "deployer"
}
