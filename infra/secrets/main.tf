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

resource "azurerm_key_vault_secret" "azure_storage_account_name" {
  name         = "azure-storage-account-name"
  value        = data.azurerm_storage_account.main.primary_access_key
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "azure_storage_account_key" {
  name         = "azure-storage-account-key"
  value        = data.azurerm_storage_account.main.primary_access_key
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "storage_account_blob_endpoint" {
  name         = "azure-storage-blob-endpoint"
  value        = data.azurerm_storage_account.main.primary_blob_endpoint
  key_vault_id = data.azurerm_key_vault.main.id
}

# ------ PostgreSQL Flexible Server Secrets ------
resource "azurerm_key_vault_secret" "postgress_password" {
  name         = "postgres-password"
  value        = var.postgres_password
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "postgress_host" {
  name         = "postgres-host"
  value        = var.postgres_server_name
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "postgres_user" {
  name         = "postgres-user"
  value        = var.postgres_user
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "postgres_database" {
  name         = "postgres-database"
  value        = var.postgres_database
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
  name                = var.redis_hostname
  resource_group_name = var.resource_group_name
}

resource "azurerm_key_vault_secret" "redis_hostname" {
  name         = "redis-hostname"
  value        = var.redis_hostname
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "redis_password" {
  name         = "redis-password"
  value        = data.azurerm_managed_redis.main.default_database[0].primary_access_key
  key_vault_id = data.azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "redis_url" {
  name         = "redis-url"
  value        = "rediss://${var.redis_hostname}:10000"
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

resource "azurerm_key_vault_secret" "rabbitmq_default_user" {
  name         = "rabbitmq-default-user"
  value        = var.rabbitmq_default_user
  key_vault_id = data.azurerm_key_vault.main.id
}
