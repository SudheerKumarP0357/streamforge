output "resource_group_name" {
  value = azurerm_resource_group.main.name
}

output "storage_account_name" {
  value = azurerm_storage_account.main.name
}

output "key_vault_name" {
  value = azurerm_key_vault.main.name
}

output "postgres_server_name" {
  value = azurerm_postgresql_flexible_server.main.fqdn
}

output "redis_name" {
  value = azurerm_managed_redis.main.name
}

output "aks_cluster_name" {
  value = azurerm_kubernetes_cluster.main.name
}

output "app_vnet_name" {
  value = azurerm_virtual_network.app_vnet.name
}

output "mongo_cluser_name" {
  value = azurerm_mongo_cluster.main.name
}
