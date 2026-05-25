resource "azurerm_nat_gateway" "main" {
  name                    = "nat-sf${local.name}"
  location                = azurerm_resource_group.main.location
  resource_group_name     = azurerm_resource_group.main.name
  sku_name                = "StandardV2"
  idle_timeout_in_minutes = 4
}

resource "azurerm_public_ip" "nat_pip" {
  name                = "nat-pip-sf${local.name}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "StandardV2"
  allocation_method   = "Static"
}

resource "azurerm_nat_gateway_public_ip_association" "main" {
  nat_gateway_id       = azurerm_nat_gateway.main.id
  public_ip_address_id = azurerm_public_ip.nat_pip.id
}

resource "azurerm_subnet_nat_gateway_association" "aks_nat" {
  subnet_id      = azurerm_subnet.aks_subnet.id
  nat_gateway_id = azurerm_nat_gateway.main.id
}
