# resource "azurerm_network_security_group" "aks_nsg" {
#   name                = "aks-nsg-${local.name}"
#   location            = azurerm_resource_group.main.location
#   resource_group_name = azurerm_resource_group.main.name
#   tags                = var.tags
# }

# resource "azurerm_subnet_network_security_group_association" "aks_nsg" {
#   subnet_id                 = azurerm_subnet.aks_subnet.id
#   network_security_group_id = azurerm_network_security_group.aks_nsg.id
# }
