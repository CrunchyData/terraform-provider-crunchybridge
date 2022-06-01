data "crunchybridge_clusterids" "lookup" {}

output "cluster_ids" {
  value = data.crunchybridge_clusterids.lookup
}
