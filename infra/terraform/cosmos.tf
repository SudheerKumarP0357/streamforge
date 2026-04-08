resource "azurerm_mongo_cluster" "main" {
  name                = "cosmon-${var.application_name}-${var.environment_name}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  administrator_username = var.cosmos_administrator_username
  administrator_password = var.cosmos_administrator_password

  compute_tier = var.cosmos_compute_tier

  high_availability_mode = "Disabled"
  shard_count            = "1"
  storage_size_in_gb     = "32"
  version                = "8.0"
  public_network_access  = "Disabled"

  tags = var.tags
}

module "cosmos" {
  source                = "./modules/private-endpoint"
  resource_group_name   = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  subnet_id             = azurerm_subnet.cosmos_subnet.id
  private_endpoint_name = "cosmos"
  resource_id           = azurerm_mongo_cluster.main.id
  subresource_names     = ["MongoCluster"]
  tags                  = var.tags
  private_dns_zone_name = "privatelink.mongocluster.cosmos.azure.com"
  jump_vnet_id          = azurerm_virtual_network.jump_vnet.id
  app_vnet_id           = azurerm_virtual_network.app_vnet.id
}
