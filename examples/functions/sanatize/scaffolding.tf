terraform {
  required_providers {
    terraconf = {
      source  = "registry.terraform.io/rodneymkay/terraconf"
      version = "1.0.0"
    }
  }
}

provider "terraconf" {
}
