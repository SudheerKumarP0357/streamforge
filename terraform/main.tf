resource "azurerm_resource_group" "main" {
  name     = "rg-${var.application_name}-${var.environment_name}-${var.primary_location_short_name}"
  location = var.primary_location
  tags     = var.tags
}
