application_name            = "streamforge"
environment_name            = "test"
primary_location            = "centralindia"
primary_location_short_name = "ci"
vnet_app_addr_space         = "10.10.0.0/16"
vnet_jump_addr_space        = "10.9.0.0/16"

# cosmos
cosmos_administrator_username = "streamforge"
cosmos_compute_tier           = "Free"

# postgres
postgres_administrator_username = "streamforge"
postgres_server_version         = 17
postgres_storage_in_mb          = 32768
postgres_storage_tier           = "P4"
postgres_sku                    = "B_Standard_B1ms"
postgres_database_name          = "streamforge"
postgres_extensions             = "PGCRYPTO,UUID-OSSP"


# Managed Redis
redis_sku_name = "Balanced_B0"


# Key Vault
key_vault_admin_object_id = "71585d88-7d50-4edc-b591-52f60aa5de92"


# ACR
acr_sku = "Basic"

# AKS
aks_kubernetes_version           = "1.35.0"
aks_sku_tier                     = "Free"
aks_max_system_pool_node_count   = 1
aks_min_system_pool_node_count   = 1
aks_systempool_max_pods_per_node = 250
kubectl_version                  = "1.35"
