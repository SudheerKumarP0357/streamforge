resource "azurerm_role_assignment" "main" {
  principal_id         = var.principal_id
  principal_type       = var.principal_type
  role_definition_name = var.role_definition_name
  scope                = var.scope
}
