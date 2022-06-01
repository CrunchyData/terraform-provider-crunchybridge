data "crunchybridge_cluster" "demo" {
  id = var.example_id
}

output "demo" {
  value = data.crunchybridge_cluster.demo
}
