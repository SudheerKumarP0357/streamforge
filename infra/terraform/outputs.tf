output "resource_group_name" {
  value = azurerm_resource_group.main.name
}

output "storage_account_name" {
  value = azurerm_storage_account.main.name
}

output "key_vault_name" {
  value = azurerm_key_vault.main.name
}

output "storage_account_blob_endpoint" {
  value = azurerm_storage_account.main.primary_blob_endpoint
}

output "postgres_server_name" {
  value = azurerm_postgresql_flexible_server.main.fqdn
}

output "postgres_password" {
  value     = azurerm_postgresql_flexible_server.main.administrator_password
  sensitive = true
}

output "postgres_database_name" {
  value = var.postgres_database_name
}

output "redis_hostname" {
  value = azurerm_managed_redis.main.hostname
}

output "redis_name" {
  value = azurerm_managed_redis.main.name
}

output "cosmos_cluster_connection_string" {
  value = [
    for cs in azurerm_mongo_cluster.main.connection_strings : cs.value if cs.name == "GlobalReadWrite"
  ][0]
  sensitive = true
}

# output "acr_login_server" {
#   value = azurerm_container_registry.acr.login_server
# }

output "aks_cluster_name" {
  value = azurerm_kubernetes_cluster.main.name
}

output "workload_identity_client_id" {
  value = azurerm_user_assigned_identity.sf_workload_identity.client_id
}

# output "jump_server_user_name" {
#   value = azurerm_linux_virtual_machine.jump_server.admin_username
# }

# output "jump_server_pip" {
#   value = azurerm_linux_virtual_machine.jump_server.public_ip_address
# }

output "alb_subnet_id" {
  value = azurerm_subnet.alb_subnet.id
}
output "azure_alb_identity" {
  value = azurerm_user_assigned_identity.alb_uami.client_id
}
