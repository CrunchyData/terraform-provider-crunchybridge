data "crunchybridge_clusterstatus" "demostatus" {
  id = var.example_id
}

output "status" {
  value = data.crunchybridge_clusterstatus.demostatus
}
