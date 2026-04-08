resource "azurerm_public_ip" "jump_server" {
  name                = "pip-jumpserver"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  allocation_method   = "Static"

  tags = var.tags
}

resource "azurerm_network_interface" "jump_server" {
  name                = "jumpserver-nic"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  ip_configuration {
    name                          = "public"
    subnet_id                     = azurerm_subnet.jump_subnet.id
    public_ip_address_id          = azurerm_public_ip.jump_server.id
    private_ip_address_allocation = "Dynamic"
  }

  tags = var.tags
}

resource "azurerm_network_security_group" "jump_server" {
  name                = "nsg-jumpserver"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = var.tags
}

resource "azurerm_network_security_rule" "jump_server_ssh_rule" {
  name                        = "AllowSSHFromMyIP"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = chomp(data.http.my_ip.response_body)
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.main.name
  network_security_group_name = azurerm_network_security_group.jump_server.name
}

resource "azurerm_subnet_network_security_group_association" "jump_server" {
  subnet_id                 = azurerm_subnet.jump_subnet.id
  network_security_group_id = azurerm_network_security_group.jump_server.id
}

resource "azurerm_linux_virtual_machine" "jump_server" {
  name                = "jumphost"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_B1ms"
  admin_username      = "adminuser"

  custom_data = base64encode(templatefile("${path.module}/cloud-init.yml", {
    aks_cluster_name    = "${var.application_name}${var.environment_name}${var.primary_location_short_name}"
    resource_group_name = azurerm_resource_group.main.name
    kubectl_version     = var.kubectl_version
  }))

  network_interface_ids = [
    azurerm_network_interface.jump_server.id,
  ]

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.jump_server.id]
  }

  admin_ssh_key {
    username = "adminuser"
    # public_key = tls_private_key.jump_server_ssh_keys.public_key_openssh
    public_key = var.public_key_openssh
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "ubuntu-24_04-lts"
    sku       = "server"
    version   = "latest"
  }

  tags = var.tags
}


resource "azurerm_user_assigned_identity" "jump_server" {
  name                = "jumpserver-uami"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  tags                = var.tags
}


### BELOW ASSIGMENTS ARE NEEDED FOR JUMPSERVER TO ACCESS AKS CLUSTER AND UPDATE WITH ALB-CONTROLLER ADD-ON
resource "azurerm_role_assignment" "aks_cluster_user_role" {
  principal_id         = azurerm_user_assigned_identity.jump_server.principal_id
  role_definition_name = "Azure Kubernetes Service Cluster User Role"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_kubernetes_cluster.main.id
}

resource "azurerm_role_assignment" "aks_cluster_contributor_role" {
  principal_id         = azurerm_user_assigned_identity.jump_server.principal_id
  role_definition_name = "Azure Kubernetes Service Contributor Role"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_kubernetes_cluster.main.id
}

resource "azurerm_role_assignment" "vnet_app_network_contributor" {
  principal_id         = azurerm_user_assigned_identity.jump_server.principal_id
  role_definition_name = "Network Contributor"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_virtual_network.app_vnet.id
}

resource "azurerm_role_assignment" "sf_aks_uami_mio" {
  principal_id         = azurerm_user_assigned_identity.jump_server.principal_id
  role_definition_name = "Managed Identity Operator"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_user_assigned_identity.sf_aks_uami.id
}

resource "azurerm_role_assignment" "aks_node_resource_group_reader" {
  principal_id         = azurerm_user_assigned_identity.jump_server.principal_id
  role_definition_name = "Reader"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_kubernetes_cluster.main.node_resource_group_id
}
### ABOVE ASSIGMENTS ARE NEEDED FOR JUMPSERVER TO ACCESS AKS CLUSTER AND UPDATE WITH ALB-CONTROLLER ADD-ON
