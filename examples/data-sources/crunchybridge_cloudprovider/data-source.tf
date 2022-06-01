data "crunchybridge_cloudprovider" "aws" {
  provider_id = "aws"
}

output "plans" {
  value = data.crunchybridge_cloudprovider.aws.plans
}

output "regions" {
  value = data.crunchybridge_cloudprovider.aws.regions
}
