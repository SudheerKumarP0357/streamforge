data "http" "my_ip" {
  url = "https://ifconfig.me/ip"
}

data "azurerm_key_vault" "main" {
  name                = var.key_vault_name
  resource_group_name = var.resource_group_name
}


# ------ Storage Account Secrets ------
data "azurerm_storage_account" "main" {
  name                = var.storage_account_name
  resource_group_name = var.resource_group_name
}

resource "azurerm_key_vault_secret" "azure_storage_account_key" {
  name         = "azure-storage-account-key"
  value        = data.azurerm_storage_account.main.primary_access_key
  key_vault_id = data.azurerm_key_vault.main.id
}

# ------ PostgreSQL Flexible Server Secrets ------
resource "azurerm_key_vault_secret" "postgress_password" {
  name         = "postgres-password"
  value        = var.postgres_administrator_password
  key_vault_id = data.azurerm_key_vault.main.id
}


# ------ Cosmos DB Secrets ------
resource "azurerm_key_vault_secret" "cosmos_connection_string" {
  name         = "cosmos-connection-string"
  value        = var.cosmos_cluster_connection_string
  key_vault_id = data.azurerm_key_vault.main.id
}


# ------ Redis Secrets ------
data "azurerm_managed_redis" "main" {
  name                = var.redis_name
  resource_group_name = var.resource_group_name
}
resource "azurerm_key_vault_secret" "redis_password" {
  name         = "redis-password"
  value        = data.azurerm_managed_redis.main.default_database[0].primary_access_key
  key_vault_id = data.azurerm_key_vault.main.id
}


# ------ JWT Secret ------
resource "azurerm_key_vault_secret" "jwt_secret" {
  name         = "jwt-secret"
  value        = var.jwt_secret
  key_vault_id = data.azurerm_key_vault.main.id
}


# ------ RabbitMQ Secrets ------
resource "azurerm_key_vault_secret" "rabbitmq_default_pass" {
  name         = "rabbitmq-default-pass"
  value        = var.rabbitmq_default_pass
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "rabbitmq_url" {
  name         = "rabbitmq-url"
  value        = "amqp://${var.rabbitmq_default_user}:${var.rabbitmq_default_pass}@sf-rabbitmq-svc:5672/"
  key_vault_id = data.azurerm_key_vault.main.id
}
