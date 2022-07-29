---
subcategory: ""
page_title: "Refactoring Terraform Variables"
description: |-
    An example of how to separate Terraform variable definitions from module files.
---

# Refactoring variables

This guide is a quick reference to bridge the guides' use of input variables. It shows a result of refactoring of the quick example to use input variables to separate sensitive configuration values from the module content. This is particularly useful when committing a module to source code management systems.

The Terraform team maintains documentation on their configuration language, including [Declaring an Input Variable](https://www.terraform.io/language/values/variables#declaring-an-input-variable).

`insensitive-module.tf`
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
}

data "crunchybridge_clusterstatus" "status" {
  id = crunchybridge_cluster.demo.id
}

output "demo_status" {
  value = data.crunchybridge_clusterstatus.status
}
```

`variables.tf`
```terraform
variable "api_key" {
  type        = string
  description = "Crunchy Bridge API key - Application ID"
}

variable "api_secret" {
  type        = string
  description = "Crunchy Bridge API key - Application Secret"
  sensitive   = true
}

variable "example_id" {
  type        = string
  description = "Cluster ID to be used in example .tf files"
}
```

The following file would not be stored as part of the module's source in a repository and may be generated per-environment.

`terraform.tfvars`
```terraform
// same invalid creds as quick example
api_key    = "i72j3uzffbglzbhfrhnalgirwq"
api_secret = "uLXm2Icjsv2w-8kWu3g0ToEaM0mDqbDk9v-AjFkzgD4"
example_id = "prir3viwrfgy5dkb5ddiwb4nsu"
```
