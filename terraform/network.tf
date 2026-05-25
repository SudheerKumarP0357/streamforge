resource "azurerm_virtual_network" "app_vnet" {
  name                = "vnet-app-${local.name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = var.primary_location
  address_space       = [var.vnet_app_addr_space]

  tags = var.tags
}

resource "azurerm_virtual_network" "jump_vnet" {
  name                = "vnet-jump-${local.name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = var.primary_location
  address_space       = [var.vnet_jump_addr_space]

  tags = var.tags
}


# ===============================================================
# JUMP VNET & APP VNET PEERINGS
# ===============================================================

resource "azurerm_virtual_network_peering" "jump_to_app" {
  name                      = "peer-jump-to-app"
  resource_group_name       = azurerm_resource_group.main.name
  virtual_network_name      = azurerm_virtual_network.jump_vnet.name
  remote_virtual_network_id = azurerm_virtual_network.app_vnet.id
}

resource "azurerm_virtual_network_peering" "app_to_jump" {
  name                      = "peer-app-to-jump"
  resource_group_name       = azurerm_resource_group.main.name
  virtual_network_name      = azurerm_virtual_network.app_vnet.name
  remote_virtual_network_id = azurerm_virtual_network.jump_vnet.id
}
