resource "humanitec_artefact_version" "example" {
  name = "registry.example.com/my-frontend-app"
  ref  = "refs/heads/main"

  version = "latest"
  commit  = "75ef9faee755c70589550b513ad881e5a603182c"

  digest = "sha256:7f0b629cbb9d794b3daf19fcd686a30a039b47395545394dadc0574744996a87"
  type   = "container"
}
