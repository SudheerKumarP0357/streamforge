resource "azurerm_postgresql_flexible_server" "main" {
  name                          = "psql${var.application_name}${var.environment_name}${var.primary_location_short_name}"
  resource_group_name           = azurerm_resource_group.main.name
  location                      = azurerm_resource_group.main.location
  version                       = var.postgres_server_version
  public_network_access_enabled = false

  authentication {
    password_auth_enabled = true
  }
  administrator_login    = var.postgres_administrator_username
  administrator_password = var.postgres_administrator_password

  zone = "1"

  storage_mb   = var.postgres_storage_in_mb
  storage_tier = var.postgres_storage_tier

  sku_name = var.postgres_sku

  delegated_subnet_id = azurerm_subnet.postgres_subnet.id
  private_dns_zone_id = azurerm_private_dns_zone.postgres_dns.id

  backup_retention_days        = 7
  geo_redundant_backup_enabled = false

  tags = var.tags

  depends_on = [azurerm_private_dns_zone_virtual_network_link.vlink_postgres_app]

}

resource "azurerm_postgresql_flexible_server_database" "streamforge" {
  name      = var.postgres_database_name
  server_id = azurerm_postgresql_flexible_server.main.id
  collation = "en_US.utf8"
  charset   = "UTF8"
}

resource "azurerm_postgresql_flexible_server_configuration" "psql_extensions" {
  name      = "azure.extensions"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = var.postgres_extensions
}


resource "azurerm_private_dns_zone" "postgres_dns" {
  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = azurerm_resource_group.main.name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vlink_postgres_app" {
  name                  = "postgres-vlink-app"
  resource_group_name   = azurerm_resource_group.main.name
  private_dns_zone_name = azurerm_private_dns_zone.postgres_dns.name
  virtual_network_id    = azurerm_virtual_network.app_vnet.id
  tags                  = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vlink_postgres_jump" {
  name                  = "postgres-vlink-jump"
  resource_group_name   = azurerm_resource_group.main.name
  private_dns_zone_name = azurerm_private_dns_zone.postgres_dns.name
  virtual_network_id    = azurerm_virtual_network.jump_vnet.id
  tags                  = var.tags
}
