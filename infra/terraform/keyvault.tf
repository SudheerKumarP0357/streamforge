

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
  # network_acls {
  #   default_action = "Deny"
  #   bypass         = "AzureServices"
  #   ip_rules       = [chomp(data.http.my_ip.response_body)]
  # }

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


# NOTE: KeyVault secrets are NOT provisioned via Terraform because this
# root module runs on GitHub-hosted runners which cannot reach the
# private KeyVault data plane endpoint. Secrets are provisioned via
# az keyvault secret set in the secrets job of .github/workflows/infra.yml

# resource "azurerm_key_vault_secret" "azure_storage_account_key" {
#   name         = "azure-storage-account-key"
#   value        = azurerm_storage_account.main.primary_access_key
#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }

# resource "azurerm_key_vault_secret" "postgress_password" {
#   name         = "postgres-password"
#   value        = azurerm_postgresql_flexible_server.main.administrator_password
#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }

# resource "azurerm_key_vault_secret" "cosmos_connection_string" {
#   name  = "cosmos-connection-string"
#   value = "mongodb+srv://${var.cosmos_administrator_username}:${azurerm_mongo_cluster.main.administrator_password}@${azurerm_mongo_cluster.main.name}.global.mongocluster.cosmos.azure.com/?tls=true&authMechanism=SCRAM-SHA-256&retrywrites=false&maxIdleTimeMS=120000"

#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }

# resource "azurerm_key_vault_secret" "redis_password" {
#   name         = "redis-password"
#   value        = azurerm_managed_redis.main.default_database[0].primary_access_key
#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }

# resource "azurerm_key_vault_secret" "jwt_secret" {
#   name         = "jwt-secret"
#   value        = var.jwt_secret
#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }

# resource "azurerm_key_vault_secret" "rabbitmq_default_pass" {
#   name         = "rabbitmq-default-pass"
#   value        = var.rabbitmq_default_pass
#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }

# resource "azurerm_key_vault_secret" "rabbitmq_url" {
#   name         = "rabbitmq-url"
#   value        = "amqp://${var.rabbitmq_default_user}:${var.rabbitmq_default_pass}@sf-rabbitmq-svc:5672/"
#   key_vault_id = azurerm_key_vault.main.id

#   depends_on = [
#     azurerm_role_assignment.key_vault_admin_terraform,
#     azurerm_role_assignment.key_vault_admin
#   ]
# }




