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

output "alb_subnet_id" {
  value = azurerm_subnet.alb_subnet.id
}
output "azure_alb_identity" {
  value = azurerm_user_assigned_identity.alb_uami.client_id
}

output "mongo_cluser_name" {
  value = azurerm_mongo_cluster.main.name
}
