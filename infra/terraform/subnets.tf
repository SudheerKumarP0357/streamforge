resource "azurerm_subnet" "storage_account_subnet" {
  name                            = "snet-blob"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.10.0/24"]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "cosmos_subnet" {
  name                            = "snet-cosmos"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.11.0/24"]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "keyvault_subnet" {
  name                            = "snet-keyvault"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.12.0/24"]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "postgres_subnet" {
  name                            = "snet-postgres"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.13.0/24"]
  default_outbound_access_enabled = false

  service_endpoints = [
    "Microsoft.Storage"
  ]

  delegation {
    name = "postgres-delegation"
    service_delegation {
      name = "Microsoft.DBforPostgreSQL/flexibleServers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action"
      ]
    }
  }
}

resource "azurerm_subnet" "redis_subnet" {
  name                            = "snet-redis"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.14.0/24"]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "acr_subnet" {
  name                            = "snet-acr"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.15.0/24"]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "aks_subnet" {
  name                            = "snet-aks"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.16.0/22"]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "alb_subnet" {
  name                            = "alb-subnet"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.app_vnet.name
  address_prefixes                = ["10.10.20.0/24"]
  default_outbound_access_enabled = false

  delegation {
    name = "agfc-delegation"
    service_delegation {
      name = "Microsoft.ServiceNetworking/trafficControllers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action"
      ]
    }
  }
}


# ===============================================================
# JUMP VNET SUBNETS
# ===============================================================

resource "azurerm_subnet" "jump_subnet" {
  name                            = "snet-jump"
  resource_group_name             = azurerm_resource_group.main.name
  virtual_network_name            = azurerm_virtual_network.jump_vnet.name
  address_prefixes                = ["10.9.10.0/24"]
  default_outbound_access_enabled = false
}



# ===============================================================
# JUMP VNET & APP VNET PEERINGS
# ===============================================================

resource "azurerm_virtual_network_peering" "jump_to_app" {
  name                      = "peer-jump-to-app"
  resource_group_name       = azurerm_resource_group.main.name
  virtual_network_name      = azurerm_virtual_network.jump_vnet.name
  remote_virtual_network_id = azurerm_virtual_network.app_vnet.id
}

resource "azurerm_virtual_network_peering" "app_to_jump" {
  name                      = "peer-app-to-jump"
  resource_group_name       = azurerm_resource_group.main.name
  virtual_network_name      = azurerm_virtual_network.app_vnet.name
  remote_virtual_network_id = azurerm_virtual_network.jump_vnet.id
}
