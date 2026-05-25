# data "http" "my_ip" {
#   url = "https://ifconfig.me/ip"
# }

data "azurerm_client_config" "current" {}

data "azurerm_kubernetes_service_versions" "aks_version" {
  location        = azurerm_resource_group.main.location
  include_preview = false
}
