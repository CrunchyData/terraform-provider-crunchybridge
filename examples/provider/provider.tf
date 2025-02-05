terraform {
  required_providers {
    crunchybridge = {
      source  = "CrunchyData/crunchybridge"
      version = "0.3.0"
    }
  }
}

provider "crunchybridge" {
  application_secret = var.api_secret
}

