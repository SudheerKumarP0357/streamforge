resource "azurerm_virtual_network" "app_vnet" {
  name                = "vnet-${var.application_name}-${var.environment_name}-${var.primary_location_short_name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = var.primary_location
  address_space       = [var.vnet_app_addr_space]

  tags = var.tags
}

resource "azurerm_virtual_network" "jump_vnet" {
  name                = "vnet-jump-${var.application_name}-${var.environment_name}-${var.primary_location_short_name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = var.primary_location
  address_space       = [var.vnet_jump_addr_space]

  tags = var.tags
}
