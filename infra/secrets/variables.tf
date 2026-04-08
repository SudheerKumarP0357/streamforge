variable "resource_group_name" {
  type = string
}

variable "storage_account_name" {
  type = string
}

variable "key_vault_name" {
  type = string
}

variable "postgres_administrator_password" {
  type      = string
  sensitive = true
}

variable "cosmos_cluster_connection_string" {
  type      = string
  sensitive = true
}

variable "redis_name" {
  type = string
}

variable "jwt_secret" {
  type = string
}

variable "rabbitmq_default_pass" {
  type = string
}

variable "rabbitmq_default_user" {
  type = string
}
