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


data "template_file" "cloud_init_config" {
  template = file("${path.module}/cloud-init.yml")
}

resource "azurerm_linux_virtual_machine" "jump_server" {
  name                = "jumphost"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_B1ms"
  admin_username      = "adminuser"

  custom_data = base64encode(data.template_file.cloud_init_config.rendered)

  network_interface_ids = [
    azurerm_network_interface.jump_server.id,
  ]

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.jump_server.id]
  }

  admin_ssh_key {
    username   = "adminuser"
    public_key = file("~/.ssh/id_ed25519.pub")
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

resource "azurerm_role_assignment" "jump_server" {
  principal_id         = azurerm_user_assigned_identity.jump_server.principal_id
  role_definition_name = "Contributor"
  principal_type       = "ServicePrincipal"
  scope                = azurerm_resource_group.main.id
}
