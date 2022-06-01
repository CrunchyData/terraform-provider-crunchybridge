terraform {
  required_providers {
    crunchybridge = {
      source = "CrunchyData/crunchybridge"
    }
  }
}

provider "crunchybridge" {
  application_id     = var.api_key
  application_secret = var.api_secret
}

