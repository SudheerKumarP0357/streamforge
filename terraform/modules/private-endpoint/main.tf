resource "azurerm_private_endpoint" "main" {
  name                = "pe-${var.private_endpoint_name}"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.subnet_id

  custom_network_interface_name = "pe-${var.private_endpoint_name}-nic"

  private_service_connection {
    name                           = "pe-${var.private_endpoint_name}-connection"
    private_connection_resource_id = var.resource_id
    is_manual_connection           = false
    subresource_names              = var.subresource_names
  }
  private_dns_zone_group {
    name                 = "${var.private_endpoint_name}-dns-zone-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.main.id]
  }

  tags = var.tags
}

resource "azurerm_private_dns_zone" "main" {
  name                = var.private_dns_zone_name
  resource_group_name = var.resource_group_name
  tags                = var.tags

}

resource "azurerm_private_dns_zone_virtual_network_link" "vlink_app" {
  name                  = "vlink-${var.private_endpoint_name}-app"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.main.name
  virtual_network_id    = var.app_vnet_id

  tags = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vlink_jump" {
  name                  = "vlink-${var.private_endpoint_name}-jump"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.main.name
  virtual_network_id    = var.jump_vnet_id

  tags = var.tags
}
