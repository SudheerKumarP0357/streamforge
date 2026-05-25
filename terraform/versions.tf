terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.74.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "3.8.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "3.5.0"
    }
  }

  backend "azurerm" {}

}

provider "azurerm" {

  features {
    key_vault {
      purge_soft_delete_on_destroy          = false
      purge_soft_deleted_keys_on_destroy    = false
      purge_soft_deleted_secrets_on_destroy = false
      recover_soft_deleted_keys             = true
      recover_soft_deleted_secrets          = true
      recover_soft_deleted_key_vaults       = true
    }

    postgresql_flexible_server {
      restart_server_on_configuration_value_change = false
    }
  }
}
