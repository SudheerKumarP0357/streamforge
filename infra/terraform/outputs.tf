output "resource_group_name" {
  value = azurerm_resource_group.main.name
}

output "stroage_account_name" {
  value = azurerm_storage_account.main.name
}

output "storage_account_primary_access_key" {
  value     = azurerm_storage_account.main.primary_access_key
  sensitive = true
}

output "storage_account_blob_endpoint" {
  value = azurerm_storage_account.main.primary_blob_endpoint
}

output "raw_videos_container_name" {
  value = azurerm_storage_container.raw_videos.name
}

output "hls_videos_container_name" {
  value = azurerm_storage_container.hls_videos.name
}

output "cosmos_administrator_username" {
  value = azurerm_mongo_cluster.main.administrator_username
}

output "cosmos_mongodb_cluster_connectionstring" {
  # value     = azurerm_mongo_cluster.main.connection_strings[0].value
  value = [
    for cs in azurerm_mongo_cluster.main.connection_strings : cs.value if cs.name == "GlobalReadWrite"
  ][0]
  sensitive = true
}

output "postgres_administrator_username" {
  value = azurerm_postgresql_flexible_server.main.administrator_login
}

output "postgres_administrator_password" {
  value     = azurerm_postgresql_flexible_server.main.administrator_password
  sensitive = true
}

output "postgres_server_name" {
  value = azurerm_postgresql_flexible_server.main.fqdn
}

output "redis_hostname" {
  value = azurerm_managed_redis.main.hostname
}

output "redis_access_key" {
  value     = azurerm_managed_redis.main.default_database[0].primary_access_key
  sensitive = true
}

output "acr_username" {
  value = azurerm_container_registry.acr.admin_username
}

output "acr_login_server" {
  value = azurerm_container_registry.acr.login_server
}

output "acr_password" {
  value     = azurerm_container_registry.acr.admin_password
  sensitive = true
}

output "aks_cluser_name" {
  value = azurerm_kubernetes_cluster.main.name
}

output "workload_identity_client_id" {
  value = azurerm_user_assigned_identity.streamforge_workload_identity.client_id
}

output "jump_server_pip" {
  value = azurerm_linux_virtual_machine.jump_server.public_ip_address
}

output "jump_server_user_name" {
  value = azurerm_linux_virtual_machine.jump_server.admin_username
}
