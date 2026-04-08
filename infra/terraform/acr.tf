resource "azurerm_container_registry" "acr" {
  name                = "acr${var.application_name}${var.environment_name}${var.primary_location_short_name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = var.acr_sku
  admin_enabled       = false

  tags = var.tags
}

resource "azurerm_role_assignment" "aks_acr_pull" {
  principal_id                     = azurerm_kubernetes_cluster.main.kubelet_identity[0].object_id
  role_definition_name             = "AcrPull"
  scope                            = azurerm_container_registry.acr.id
  skip_service_principal_aad_check = true
}

# module "acr" {
#   source                = "./modules/private-endpoint"
#   resource_group_name   = azurerm_resource_group.main.name
#   location              = azurerm_resource_group.main.location
#   subnet_id             = azurerm_subnet.acr_subnet.id
#   private_endpoint_name = "acr"
#   resource_id           = azurerm_container_registry.acr.id
#   subresource_names     = ["registry"]
#   tags                  = var.tags
#   private_dns_zone_name = "privatelink.azurecr.io"
#   jump_vnet_id          = azurerm_virtual_network.jump_vnet.id
#   app_vnet_id           = azurerm_virtual_network.app_vnet.id
# }
