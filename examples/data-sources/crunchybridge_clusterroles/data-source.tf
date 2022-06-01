data "crunchybridge_clusterroles" "default" {
  id = var.example_id
}

output "superuser_uri" {
  value     = data.crunchybridge_clusterroles.default.superuser.uri
  sensitive = true
}

output "application_uri" {
  value     = data.crunchybridge_clusterroles.default.application.uri
  sensitive = true
}
