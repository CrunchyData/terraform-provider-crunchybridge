---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "crunchybridge_clusterids Data Source - terraform-provider-crunchybridge"
subcategory: ""
description: |-
  Data Source for retreiving Cluster identifiers from the user-provided label
---

# crunchybridge_clusterids (Data Source)

Data Source for retreiving Cluster identifiers from the user-provided label

## Example Usage

```terraform
data "crunchybridge_clusterids" "lookup" {}

output "cluster_ids" {
  value = data.crunchybridge_clusterids.lookup
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `team_id` (String) Limits the cluster mapping to those clusters belonging to the identified team.

### Read-Only

- `cluster_ids_by_name` (Map of String) A mapping of cluster names to their respective cluster IDs in [EID format](https://docs.crunchybridge.com/api-concepts/eid).
- `id` (String) The ID of this resource.

