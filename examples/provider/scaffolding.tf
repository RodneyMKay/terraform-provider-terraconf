terraform {
  required_providers {
    terraconf = {
      source  = "registry.terraform.io/rodneymkay/terraconf"
      version = "~> 1.0"
    }

    azurerm = {
      source  = "registry.terraform.io/hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "azurerm" {
  features {}
  resource_provider_registrations = "none"

  subscription_id = "00000000-0000-0000-0000-000000000000"
  tenant_id       = "00000000-0000-0000-0000-000000000000"
  client_id       = "00000000-0000-0000-0000-000000000000"
  client_secret   = "fake-secret"
}
