resource "azurerm_resource_group" "main" {
  name     = "rg-${local.name}"
  location = var.primary_location
  tags     = var.tags
}
