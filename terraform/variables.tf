variable "application_name" {
  type        = string
  description = "The name of the application"
}

variable "environment_name" {
  type        = string
  description = "The name of the environment (e.g., dev, prod)"
}

variable "primary_location" {
  type        = string
  description = "The primary Azure region for deploying resources"
}

variable "primary_location_short_name" {
  type        = string
  description = "A short name or abbreviation for the primary location, used in resource naming conventions"
}

variable "vnet_app_addr_space" {
  type        = string
  description = "The address space for the application virtual network"
}

variable "vnet_jump_addr_space" {
  type        = string
  description = "The address space for the jump server/bastion virtual network"
}

variable "tags" {
  type = object({
    Environment = string
    Owner       = string
    Created_By  = string
    Application = string
  })
  description = "A mapping of tags to assign to all resources"

  default = ({
    Environment = "Prod"
    Owner       = "Sudheer"
    Created_By  = "Terraform"
    Application = "Streamforge"
  })
}

variable "cosmos_administrator_username" {
  type        = string
  sensitive   = true
  description = "The administrator username for the Cosmos DB account"
}

variable "cosmos_administrator_password" {
  type        = string
  sensitive   = true
  description = "The administrator password for the Cosmos DB account"
}

variable "cosmos_compute_tier" {
  type        = string
  description = "The compute tier for Cosmos DB"
}

variable "postgres_administrator_username" {
  type        = string
  sensitive   = true
  description = "The administrator username for the PostgreSQL Flexible Server"
}

variable "postgres_administrator_password" {
  type        = string
  sensitive   = true
  description = "The administrator password for the PostgreSQL Flexible Server"
}

variable "postgres_server_version" {
  type        = number
  description = "The version of PostgreSQL to use for the Flexible Server"
}

variable "postgres_storage_in_mb" {
  type        = number
  description = "The storage capacity in megabytes for the PostgreSQL Flexible Server"
}

variable "postgres_storage_tier" {
  type        = string
  description = "The storage tier for the PostgreSQL Flexible Server"
}

variable "postgres_sku" {
  type        = string
  description = "The SKU name for the PostgreSQL Flexible Server"
}

variable "postgres_database_name" {
  type        = string
  sensitive   = true
  description = "The name of the PostgreSQL database to create"
}

variable "postgres_extensions" {
  type        = list(string)
  description = "Comma separated extensions list Ex: PGCRYPTO,PGCRYPTO2"
}

variable "redis_sku_name" {
  type        = string
  description = "The SKU name for the Azure Cache for Redis instance"
}

variable "jwt_secret" {
  type        = string
  sensitive   = true
  description = "The secret key used for signing JWT tokens"
}

variable "key_vault_admin_object_id" {
  type        = string
  description = "The object ID of the user or service principal that will be granted admin access to Key Vault"
}

variable "acr_sku" {
  type        = string
  description = "The SKU name for the Azure Container Registry"
}

variable "aks_kubernetes_version" {
  type        = string
  description = "The Kubernetes version to use for the AKS cluster"
}

variable "aks_sku_tier" {
  type        = string
  description = "The SKU tier for the AKS cluster"
}

variable "aks_min_system_pool_node_count" {
  type        = number
  description = "The minimum number of nodes for the AKS system node pool"
}

variable "aks_max_system_pool_node_count" {
  type        = number
  description = "The maximum number of nodes for the AKS system node pool"
}

variable "aks_systempool_max_pods_per_node" {
  type        = number
  description = "Maximum no of pods per node in the AKS system pool"
}

variable "kubectl_version" {
  type        = string
  description = "The version of kubectl to use"
}

# variable "public_key_openssh" {
#   type        = string
#   sensitive   = true
#   description = "The OpenSSH public key for SSH access"
# }

variable "allowed_origins" {
  type        = string
  description = "Allowed Origins for Azure Storage Account"
}
