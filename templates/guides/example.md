---
subcategory: ""
page_title: "Expanded Quick Start Example"
description: |-
    An example of how to spin up a Crunchy Bridge cluster using Terraform.
---

# Simple Working Example

This example and all guide examples assume you have already set up an account in [Crunchy Bridge](https://crunchybridge.com) and provisioned an API key in the [Account Settings](https://crunchybridge.com/account). 

The standard example usage for a Terraform provider primarily shows how to create a provider, but makes no explanation of its use.

Here is a resource creation example for the Crunchy Bridge Provider:

`one-file-example.tf`
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
  application_secret = "uLXm2Icjsv2w-8kWu3g0ToEaM0mDqbDk9v-AjFkzgD4"
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
Working credentials will need to be substituted for the `application_id` and `application_secret` examples above.

## Executing the example
Starting from a working directory containing the example above, one can execute:
```sh
> terraform init
> terraform apply
```
A Terraform user will typically run a `terraform plan` to preview the changes prior to applying them.


## Example Explanation

The `terraform` block is a typical provider import as required by Terraform documentation and includes its name `crunchybridge`, its registry source, and its version.

The `provider` block allows the configuration of the provider interface. In this case, the API key information is the minimum required configuration.


**Account data**
```terraform
data "crunchybridge_account" "user" {}
```
The first data source defined is user account data, named **user**. The account data is for the user associated with the configured API key. This data source is defined to obtain the user's default team id for the cluster creation process.

**Cluster Creation**
```terraform
resource "crunchybridge_cluster" "demo" {
  team_id = data.crunchybridge_account.user.default_team
  name    = "famously-fragile-impala-47"
}
```
The primary resource of the Crunchy Bridge Provider is the cluster. Here, it is named **demo** within Terraform and provides the minimum required fields for provisioning.

The default team id is extracted from the previously mentioned data source **user**.  
The name for the cluster was generated randomly from `tools/petname.go` in the provider repository. This tool may become a provisionable resource in the future if needed. 

**Default values**

A word of caution regarding the cluster resource's default fields - if update operations using this terraform manifest are likely, it is recommended to explicitly set the value of `plan_id`, `is_ha`, `storage`, and `major_version`. Leaving these updatable fields null in the manifest will make the Terraform update appear to be from the default value (reflected in the current status) to a null value (reflected in the manifest).

**Cluster Status**
```terraform
data "crunchybridge_clusterstatus" "status" {
  id = crunchybridge_cluster.demo.id
}

output "demo_status" {
  value = data.crunchybridge_clusterstatus.status
}
```
The default method of provisioning a cluster resource is asynchronous. The request is sent to the API and connection information is available, but the status must be checked to know when the cluster is ready for use.

In this example, the `status` data source is configured to retrieve the status of the provisioned cluster. An output named `demo_status` is set up to show the cluster status data.
Due to the asynchronous request, `terraform apply` operations can update the status output to reflect the current state of the cluster.

The Crunchy Bridge Terraform provider has a feature to perform the cluster provisioning in a seemingly synchronous manner. Setting the `wait_until_ready` parameter of the cluster resource to `true` will force the resource creation to wait until the cluster is ready for use before declaring success.

## Additional information

For brevity, this example mixes sensitive and non-sensitive configuration information. See the [variables](variables.md) guide for an example that separates sensitive variables from the manifest.

This example focuses on provisioning a cluster, but does not provide the outputs needed for connecting to the provisioned database. The [cluster roles](../data-sources/clusterroles.md) data source documentation contains an example which is used in the [service binding](service-binding.md) guide.

The [cloud providers](../data-sources/cloudprovider.md) data source gives examples that show how to identify supported plans and regions for supported infrastructure providers.
