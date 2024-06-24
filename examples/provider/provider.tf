terraform {
  required_providers {
    crunchybridge = {
      source  = "CrunchyData/crunchybridge"
      version = "0.2.0"
    }
  }
}

provider "crunchybridge" {
  application_id     = var.api_key
  application_secret = var.api_secret
}

