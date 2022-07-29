---
subcategory: ""
page_title: "Updating Cluster Resources"
description: |-
    An example of how to update a Crunchy Data resource.
---

This example starts with the resource defined in the [Quick Start Example](example.md), modified with variable definitions. It also takes the advice to make explicit those attributes which allow updates.

Some of the command-line outputs are filtered to focus on the example, so the outputs here may not directly match local executions.

`example-file.tf`
```terraform
terraform {
  required_providers {
    crunchybridge = {
      source  = "CrunchyData/crunchybridge"
      version = "0.1.0"
    }
  }
}

provider "crunchybridge" {
  application_id     = vars.api_key
  application_secret = vars.api_secret
}

data "crunchybridge_account" "user" {}

resource "crunchybridge_cluster" "demo" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "famously-fragile-impala-47"
  plan_id = "hobby-2"
  is_ha   = false
  storage = 100
  major_version    = 14
  wait_until_ready = true
}
```

After the `terraform apply` execution:

```
> terraform apply
data.crunchybridge_account.user: Reading...
data.crunchybridge_account.user: Read complete after 0s [id=t3d3wvqhjrg6rk4b6mnqq6pn6u]

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # crunchybridge_cluster.demo will be created
  + resource "crunchybridge_cluster" "demo" {
      + cpu                      = (known after apply)
      + id                       = (known after apply)
      + is_ha                    = false
      + major_version            = 14
      + memory                   = (known after apply)
      + name                     = "famously-fragile-impala-47"
      + plan_id                  = "hobby-2"
      + provider_id              = "aws"
      + region_id                = "us-west-1"
      + storage                  = 100
      + wait_until_ready         = true
    }

Plan: 1 to add, 0 to change, 0 to destroy.

crunchybridge_cluster.demo: Creating...
crunchybridge_cluster.demo: Still creating... [10s elapsed]
.
. a few minutes later
.
crunchybridge_cluster.demo: Still creating... [4m10s elapsed]
crunchybridge_cluster.demo: Still creating... [4m20s elapsed]
crunchybridge_cluster.demo: Creation complete after 4m24s [id=ccgjaruv5zf2bfreasblgo4gyu]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

Updating the resource can be performed by modifying the appropriate attributes in the manifest and reapplying.  
Most update requests are asynchronous and are performed during the defined maintenance window for the cluster. At this time, altering the maintenance window must be done using the cluster settings page within the Crunchy Bridge dashboard.

**NOTE: Without a defined maintenance window, update operations are scheduled without delay**

The following definition updates the cluster name, plan_id, and storage and adds output data to show the asynchronous output:

```terraform
resource "crunchybridge_cluster" "demo" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "calmly-robust-hippopotamus-90"
  plan_id = "hobby-4"
  is_ha   = false
  storage = 200
  major_version    = 14
  wait_until_ready = true
}

data "crunchybridge_clusterstatus" "upg_status" {
  id = crunchybridge_cluster.demo.id
}

data "crunchybridge_cluster" "cluster_status" {
  id = crunchybridge_cluster.demo.id
}

output "cluster_state" {
  value = data.crunchybridge_cluster.cluster_status
}

output "upg_state" {
  value = data.crunchybridge_clusterstatus.upg_status
}
```
**NOTE: `wait_until_ready` has no impact on update operations**
```
> terraform plan -out=upg.plan
data.crunchybridge_account.user: Reading...
data.crunchybridge_account.user: Read complete after 0s [id=t3d3wvqhjrg6rk4b6mnqq6pn6u]
crunchybridge_cluster.demo: Refreshing state... [id=ccgjaruv5zf2bfreasblgo4gyu]

Terraform used the selected providers to generate the following execution plan. 
Resource actions are indicated with the following symbols:
  ~ update in-place
 <= read (data resources)

Terraform will perform the following actions:

  # data.crunchybridge_cluster.cluster_status will be read during apply
  # (depends on a resource or a module with changes pending)
 <= data "crunchybridge_cluster" "cluster_status" {
    ...
    }

  # data.crunchybridge_clusterstatus.upg_status will be read during apply
  # (depends on a resource or a module with changes pending)
 <= data "crunchybridge_clusterstatus" "upg_status" {
    ...
    }

  # crunchybridge_cluster.demo will be updated in-place
  ~ resource "crunchybridge_cluster" "demo" {
        id                       = "ccgjaruv5zf2bfreasblgo4gyu"
      ~ name                     = "famously-fragile-impala-47" -> "calmly-robust-hippopotamus-90"
      ~ plan_id                  = "hobby-2" -> "hobby-4"
      ~ storage                  = 100 -> 200
        # (11 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.

Changes to Outputs:
  + cluster_state = {
      + cpu                      = (known after apply)
      + id                       = "ccgjaruv5zf2bfreasblgo4gyu"
      + is_ha                    = (known after apply)
      + maintenance_window_start = (known after apply)
      + memory                   = (known after apply)
      + name                     = (known after apply)
      + plan_id                  = (known after apply)
      + postgres_version_id      = (known after apply)
      + provider_id              = (known after apply)
      + region_id                = (known after apply)
      + storage                  = (known after apply)
    }
  + upg_state     = {
      + id                 = "ccgjaruv5zf2bfreasblgo4gyu"
      + operations         = (known after apply)
      + state              = (known after apply)
    }

───────────────────────────────────────────────────────────────────────────────────────────────────

Saved the plan to: upg.plan

To perform exactly these actions, run the following command to apply:
    terraform apply "upg.plan"

> terraform apply upg.plan
runchybridge_cluster.demo: Modifying... [id=ccgjaruv5zf2bfreasblgo4gyu]
crunchybridge_cluster.demo: Modifications complete after 4s [id=ccgjaruv5zf2bfreasblgo4gyu]
data.crunchybridge_clusterstatus.upg_status: Reading...
data.crunchybridge_cluster.cluster_status: Reading...
data.crunchybridge_cluster.cluster_status: Read complete after 0s [id=ccgjaruv5zf2bfreasblgo4gyu]
data.crunchybridge_clusterstatus.upg_status: Read complete after 1s [id=ccgjaruv5zf2bfreasblgo4gyu]

Apply complete! Resources: 0 added, 1 changed, 0 destroyed.

Outputs:

cluster_state = {
  "cpu" = 1
  "id" = "ccgjaruv5zf2bfreasblgo4gyu"
  "is_ha" = false
  "maintenance_window_start" = 0
  "memory" = 4
  "name" = "calmly-robust-hippopotamus-90"
  "plan_id" = "hobby-4"
  "postgres_version_id" = 14
  "provider_id" = "aws"
  "region_id" = "us-west-1"
  "storage" = 200
}
upg_state = {
  "id" = "ccgjaruv5zf2bfreasblgo4gyu"
  "operations" = tolist([
    {
      "flavor" = "resize"
      "state" = "scheduled"
    },
  ])
  "state" = "ready"
}
```
After the resize operation has completed, the outputs update to:
```
Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

cluster_state = {
  "cpu" = 1
  "id" = "ccgjaruv5zf2bfreasblgo4gyu"
  "is_ha" = false
  "maintenance_window_start" = 0
  "memory" = 4
  "name" = "calmly-robust-hippopotamus-90"
  "plan_id" = "hobby-4"
  "postgres_version_id" = 14
  "provider_id" = "aws"
  "region_id" = "us-west-1"
  "storage" = 200
}
upg_state = {
  "id" = "ccgjaruv5zf2bfreasblgo4gyu"
  "operations" = tolist(null) /* of object */
  "state" = "ready"
}

```

