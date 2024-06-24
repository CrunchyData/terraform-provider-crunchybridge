---
subcategory: ""
page_title: "Service Binding Example"
description: |-
    An example of how to provision Crunchy Data clusters and bind them to an imaginary application.
---

In this example, five resources are provisioned with five data sources to retrieve the connection strings for the imaginary web application **monitor**.

`binding-example.tf`
```terraform
terraform {
  required_providers {
    crunchybridge = {
      source = "CrunchyData/crunchybridge"
      version = "0.2.0"
    }
  }
}

provider "crunchybridge" {
  application_secret = vars.app_secret
}

data "crunchybridge_account" "user" {}

resource "crunchybridge_cluster" "earth-616" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "earth-616"
}

resource "crunchybridge_cluster" "earth-838" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "earth-838"
}

resource "crunchybridge_cluster" "earth-1218" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "earth-1218"
}

resource "crunchybridge_cluster" "earth-1610" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "earth-1610"
}

resource "crunchybridge_cluster" "earth-199999" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "earth-199999"
}

data "crunchybridge_clusterroles" "earth-616" {
  id = crunchybridge_cluster.earth-616.id
}

data "crunchybridge_clusterroles" "earth-838" {
  id = crunchybridge_cluster.earth-838.id
}

data "crunchybridge_clusterroles" "earth-1218" {
  id = crunchybridge_cluster.earth-1218.id
}

data "crunchybridge_clusterroles" "earth-1610" {
  id = crunchybridge_cluster.earth-1610.id
}

data "crunchybridge_clusterroles" "earth-199999" {
  id = crunchybridge_cluster.earth-199999.id
}

// Since this provider is fake, there is no corresponding import/provider block for it
resource "fakeprovider_watcher" "monitor" {
  watch_targets = [
    data.crunchybridge_clusterroles.earth-616.application.uri,
    data.crunchybridge_clusterroles.earth-838.application.uri,
    data.crunchybridge_clusterroles.earth-1218.application.uri,
    data.crunchybridge_clusterroles.earth-1610.application.uri,
    data.crunchybridge_clusterroles.earth-199999.application.uri
  ]
}

```
