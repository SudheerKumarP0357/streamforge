terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.65.0"
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
      purge_soft_delete_on_destroy          = true  # hard-delete the vault on destroy
      purge_soft_deleted_keys_on_destroy    = true  # hard-delete keys inside it
      purge_soft_deleted_secrets_on_destroy = true  # hard-delete secrets inside it
      recover_soft_deleted_keys             = false # don't auto-recover keys
      recover_soft_deleted_secrets          = false # don't auto-recover secrets
      recover_soft_deleted_key_vaults       = false
    }

    postgresql_flexible_server {
      restart_server_on_configuration_value_change = false
    }
  }
}

