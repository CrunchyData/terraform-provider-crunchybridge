data "crunchybridge_account" "user" {}

resource "crunchybridge_cluster" "demo" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "famously-fragile-impala-47"
}

data "crunchybridge_clusterstatus" "status" {
  id = crunchybridge_cluster.demo.id
}

output "demo_status" {
  value = data.crunchybridge_clusterstatus.status
}
