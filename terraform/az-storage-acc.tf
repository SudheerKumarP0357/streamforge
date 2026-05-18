resource "azurerm_storage_account" "main" {
  name                     = "st${var.application_name}${var.environment_name}${var.primary_location_short_name}"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  account_kind               = "StorageV2"
  is_hns_enabled             = true
  https_traffic_only_enabled = true
  min_tls_version            = "TLS1_2"

  allow_nested_items_to_be_public = false
  public_network_access_enabled   = false
  local_user_enabled              = false



  blob_properties {

    cors_rule {
      allowed_origins    = ["https://streamforge.sudheer.fun"]
      allowed_methods    = ["GET"]
      allowed_headers    = ["*"]
      exposed_headers    = ["*"]
      max_age_in_seconds = 84600
    }

    delete_retention_policy {
      days                     = "7"
      permanent_delete_enabled = true
    }
    container_delete_retention_policy {
      days = "7"
    }
  }

  tags = var.tags

}

resource "azurerm_storage_container" "raw_videos" {
  name                  = "raw-videos"
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"
}

resource "azurerm_storage_container" "hls_videos" {
  name                  = "hls-videos"
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"
}


module "blob" {
  source                = "./modules/private-endpoint"
  resource_group_name   = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  subnet_id             = azurerm_subnet.storage_account_subnet.id
  private_endpoint_name = "blob"
  resource_id           = azurerm_storage_account.main.id
  subresource_names     = ["blob"]
  tags                  = var.tags
  private_dns_zone_name = "privatelink.blob.core.windows.net"
  jump_vnet_id          = azurerm_virtual_network.jump_vnet.id
  app_vnet_id           = azurerm_virtual_network.app_vnet.id
}
