data "crunchybridge_account" "myacct" {}

output "my_account" {
  value = data.crunchybridge_account.myacct
}
