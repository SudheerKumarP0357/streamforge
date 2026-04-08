variable "application_name" {
  type = string
}

variable "environment_name" {
  type = string
}

variable "primary_location" {
  type = string
}

variable "primary_location_short_name" {
  type = string
}

variable "vnet_app_addr_space" {
  type = string
}

variable "vnet_jump_addr_space" {
  type = string
}

variable "tags" {
  type = object({
    Environment = string
    Owner       = string
    Created_By  = string
    Application = string
  })

  default = ({
    Environment = "dev"
    Owner       = "Sudheer"
    Created_By  = "Terraform"
    Application = "Streamforge"
  })
}

variable "cosmos_administrator_username" {
  type = string
}

variable "cosmos_administrator_password" {
  type = string
}

variable "cosmos_compute_tier" {
  type = string
}

variable "postgres_administrator_username" {
  type = string
}

variable "postgres_administrator_password" {
  type = string
}

variable "postgres_server_version" {
  type = number
}

variable "postgres_storage_in_mb" {
  type = number
}

variable "postgres_storage_tier" {
  type = string
}

variable "postgres_sku" {
  type = string
}

variable "postgres_database_name" {
  type = string
}

variable "postgres_extensions" {
  type        = string
  description = "Comma separated extensions list Ex: PGCRYPTO,PGCRYPTO2"
}

variable "redis_sku_name" {
  type = string
}

variable "jwt_secret" {
  type = string
}

variable "rabbitmq_default_user" {
  type = string
}

variable "rabbitmq_default_pass" {
  type = string
}

variable "key_vault_admin_object_id" {
  type = string
}

variable "acr_sku" {
  type = string
}

variable "aks_kubernetes_version" {
  type = string
}

variable "aks_sku_tier" {
  type = string
}

variable "aks_min_system_pool_node_count" {
  type = number
}

variable "aks_max_system_pool_node_count" {
  type = number
}

variable "aks_systempool_max_pods_per_node" {
  type        = number
  description = "Maximum no of pods per node"
}

variable "kubectl_version" {
  type = string
}

variable "public_key_openssh" {
  type = string
}
