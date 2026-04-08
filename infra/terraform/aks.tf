resource "azurerm_kubernetes_cluster" "main" {
  name                              = "${var.application_name}${var.environment_name}${var.primary_location_short_name}"
  location                          = azurerm_resource_group.main.location
  resource_group_name               = azurerm_resource_group.main.name
  dns_prefix                        = var.application_name
  kubernetes_version                = var.aks_kubernetes_version
  role_based_access_control_enabled = true
  private_cluster_enabled           = true
  workload_identity_enabled         = true
  oidc_issuer_enabled               = true

  image_cleaner_enabled        = true
  image_cleaner_interval_hours = 168

  automatic_upgrade_channel = "patch"
  sku_tier                  = var.aks_sku_tier

  private_cluster_public_fqdn_enabled = false
  http_application_routing_enabled    = false
  azure_policy_enabled                = false
  local_account_disabled              = false

  private_dns_zone_id = azurerm_private_dns_zone.aks.id

  default_node_pool {
    auto_scaling_enabled    = true
    min_count               = var.aks_min_system_pool_node_count
    max_count               = var.aks_max_system_pool_node_count
    name                    = "systempool"
    vm_size                 = "Standard_D2as_v4"
    os_sku                  = "Ubuntu"
    max_pods                = var.aks_systempool_max_pods_per_node
    vnet_subnet_id          = azurerm_subnet.aks_subnet.id
    os_disk_type            = "Ephemeral"
    os_disk_size_gb         = 64
    host_encryption_enabled = true
    upgrade_settings {
      max_surge = "10%"
    }
  }

  key_vault_secrets_provider {
    secret_rotation_enabled  = true
    secret_rotation_interval = "2m"
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.sf_aks_uami.id]
  }

  network_profile {
    network_data_plane  = "azure"
    network_plugin_mode = "overlay"
    network_plugin      = "azure"
    network_policy      = "calico"
    service_cidr        = "10.40.0.0/16"
    dns_service_ip      = "10.40.0.10"
    load_balancer_sku   = "standard"
  }

  node_os_upgrade_channel = "NodeImage"
  maintenance_window_auto_upgrade {
    day_of_month = 0
    day_of_week  = "Sunday"
    duration     = 8
    frequency    = "Weekly"
    interval     = 1
    start_date   = "2026-03-29T00:00:00Z"
    start_time   = "00:00"
    utc_offset   = "+00:00"
  }
  maintenance_window_node_os {
    day_of_month = 0
    day_of_week  = "Sunday"
    duration     = 8
    frequency    = "Weekly"
    interval     = 1
    start_date   = "2026-03-29T00:00:00Z"
    start_time   = "00:00"
    utc_offset   = "+00:00"
  }

  tags = var.tags

  depends_on = [
    module.sf_aks_uami_network_contributor,
    module.sf_aks_uami_dns_contributor,
  ]

}

resource "azurerm_private_dns_zone" "aks" {
  name                = "privatelink.${var.primary_location}.azmk8s.io"
  resource_group_name = azurerm_resource_group.main.name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vlink_aks_app" {
  name                  = "vlink-aks-app"
  resource_group_name   = azurerm_resource_group.main.name
  private_dns_zone_name = azurerm_private_dns_zone.aks.name
  virtual_network_id    = azurerm_virtual_network.app_vnet.id
  tags                  = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vlink_aks_jump" {
  name                  = "vlink-aks-jump"
  resource_group_name   = azurerm_resource_group.main.name
  private_dns_zone_name = azurerm_private_dns_zone.aks.name
  virtual_network_id    = azurerm_virtual_network.jump_vnet.id
  tags                  = var.tags
}

# Needed for Private Cluster - Default User Assigned Identity
resource "azurerm_user_assigned_identity" "sf_aks_uami" {
  location            = azurerm_resource_group.main.location
  name                = "${var.application_name}-aks-uami"
  resource_group_name = azurerm_resource_group.main.name
  tags                = var.tags
}

# Needed for Private Cluster - Default User Assigned Identity Role Assignments
module "sf_aks_uami_network_contributor" {
  source               = "./modules/role_assignments"
  principal_id         = azurerm_user_assigned_identity.sf_aks_uami.principal_id
  principal_type       = "ServicePrincipal"
  role_definition_name = "Network Contributor"
  scope                = azurerm_virtual_network.app_vnet.id
}

module "sf_aks_uami_dns_contributor" {
  source               = "./modules/role_assignments"
  principal_id         = azurerm_user_assigned_identity.sf_aks_uami.principal_id
  principal_type       = "ServicePrincipal"
  role_definition_name = "Private DNS Zone Contributor"
  scope                = azurerm_private_dns_zone.aks.id
}


# Needed for Workload Identity - Used for Integrating with KeyVault CSI
resource "azurerm_user_assigned_identity" "sf_workload_identity" {
  location            = azurerm_resource_group.main.location
  name                = "${var.application_name}-workload-uami"
  resource_group_name = azurerm_resource_group.main.name
  tags                = var.tags
}

resource "azurerm_federated_identity_credential" "sf_workload_identity_fic" {
  name                      = "${azurerm_user_assigned_identity.sf_workload_identity.name}-fic"
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.main.oidc_issuer_url
  user_assigned_identity_id = azurerm_user_assigned_identity.sf_workload_identity.id
  subject                   = "system:serviceaccount:streamforge:sf-workload-sa"
}

module "sf_workload_identity_kv_reader" {
  source               = "./modules/role_assignments"
  principal_id         = azurerm_user_assigned_identity.sf_workload_identity.principal_id
  role_definition_name = "Key Vault Secrets User"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_key_vault.main.id
}


# Needed for Azure Load Balancer - Azure ALB Controller
resource "azurerm_user_assigned_identity" "alb_uami" {
  name                = "azure-alb-identity"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  tags                = var.tags
}

module "alb_reader_access_aks_mc" {
  source               = "./modules/role_assignments"
  principal_id         = azurerm_user_assigned_identity.alb_uami.principal_id
  scope                = azurerm_kubernetes_cluster.main.node_resource_group_id
  principal_type       = "ServicePrincipal"
  role_definition_name = "Reader"
}

module "alb_appw_config_manager" {
  source               = "./modules/role_assignments"
  principal_id         = azurerm_user_assigned_identity.alb_uami.principal_id
  scope                = azurerm_kubernetes_cluster.main.node_resource_group_id
  principal_type       = "ServicePrincipal"
  role_definition_name = "AppGW for Containers Configuration Manager"
}

module "alb_network_contributor" {
  source               = "./modules/role_assignments"
  principal_id         = azurerm_user_assigned_identity.alb_uami.principal_id
  scope                = azurerm_subnet.alb_subnet.id
  principal_type       = "ServicePrincipal"
  role_definition_name = "Network Contributor"
}

resource "azurerm_federated_identity_credential" "alb_fic" {
  name                      = azurerm_user_assigned_identity.alb_uami.name
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.main.oidc_issuer_url
  subject                   = "system:serviceaccount:azure-alb-system:alb-controller-sa"
  user_assigned_identity_id = azurerm_user_assigned_identity.alb_uami.id
}
