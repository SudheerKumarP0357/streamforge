

resource "azurerm_key_vault" "main" {
  name                        = "kv-${var.application_name}${var.environment_name}${var.primary_location_short_name}"
  location                    = azurerm_resource_group.main.location
  resource_group_name         = azurerm_resource_group.main.name
  enabled_for_disk_encryption = false
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  rbac_authorization_enabled = true

  public_network_access_enabled = false

  sku_name = "standard"

  tags = var.tags
}

resource "azurerm_role_assignment" "key_vault_admin" {
  scope                = azurerm_key_vault.main.id
  principal_id         = var.key_vault_admin_object_id
  role_definition_name = "Key Vault Administrator"
}

resource "azurerm_role_assignment" "key_vault_admin_terraform" {
  scope                = azurerm_key_vault.main.id
  principal_id         = data.azurerm_client_config.current.object_id
  role_definition_name = "Key Vault Administrator"
}

module "keyvault" {
  source                = "./private-endpoint"
  resource_group_name   = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  subnet_id             = azurerm_subnet.keyvault_subnet.id
  private_endpoint_name = "keyvault"
  resource_id           = azurerm_key_vault.main.id
  subresource_names     = ["Vault"]
  tags                  = var.tags
  private_dns_zone_name = "privatelink.vaultcore.azure.net"
  jump_vnet_id          = azurerm_virtual_network.jump_vnet.id
  app_vnet_id           = azurerm_virtual_network.app_vnet.id
}
