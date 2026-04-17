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
    # helm = {
    #   source  = "hashicorp/helm"
    #   version = "3.1.1"
    # }
    # kubernetes = {
    #   source  = "hashicorp/kubernetes"
    #   version = "3.0.1"
    # }
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
      recover_soft_deleted_key_vaults       = false # don't auto-recover key vaults
    }

    postgresql_flexible_server {
      restart_server_on_configuration_value_change = false
    }
  }
}

# COMMENTED SINCE HELM & KUBERNETES ARE NOT SUPPORTED UNLESS CLUSTER AND KUBE_CONFIG IS NOT AVAILABLE
# USING DIFFERENT PIPELINE TO SETUP/PREPARE THE CLUSTER
# provider "helm" {
#   kubernetes = {
#     host = azurerm_kubernetes_cluster.main.kube_config.0.host

#     client_certificate     = base64decode(azurerm_kubernetes_cluster.main.kube_config.0.client_certificate)
#     client_key             = base64decode(azurerm_kubernetes_cluster.main.kube_config.0.client_key)
#     cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.main.kube_config.0.cluster_ca_certificate)
#   }
# }


# provider "kubernetes" {
#   host = azurerm_kubernetes_cluster.main.kube_config.0.host

#   client_certificate     = base64decode(azurerm_kubernetes_cluster.main.kube_config.0.client_certificate)
#   client_key             = base64decode(azurerm_kubernetes_cluster.main.kube_config.0.client_key)
#   cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.main.kube_config.0.cluster_ca_certificate)
# }
