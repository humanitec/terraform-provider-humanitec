resource "humanitec_user" "service_user" {
  name = "example-service-user"
  role = "administrator"
  type = "service"
}

resource "humanitec_service_user_token" "token" {
  id          = "example-service-token"
  user_id     = humanitec_user.service_user.id
  description = "example token description"
  expires_at  = "2024-04-23T21:59:59.999Z"
}
