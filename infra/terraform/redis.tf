resource "azurerm_managed_redis" "main" {
  name                = "amr${var.application_name}${var.environment_name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku_name            = var.redis_sku_name

  public_network_access = "Disabled"
  default_database {
    access_keys_authentication_enabled = true
  }
  tags = var.tags
}

module "redis" {
  source                = "./private-endpoint"
  resource_group_name   = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  subnet_id             = azurerm_subnet.redis_subnet.id
  private_endpoint_name = "redis"
  resource_id           = azurerm_managed_redis.main.id
  subresource_names     = ["redisEnterprise"]
  tags                  = var.tags
  private_dns_zone_name = "privatelink.redis.azure.net"
  jump_vnet_id          = azurerm_virtual_network.jump_vnet.id
  app_vnet_id           = azurerm_virtual_network.app_vnet.id
}
