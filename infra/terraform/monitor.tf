resource "azurerm_log_analytics_workspace" "main" {
  name                = "log-${var.application_name}-${var.environment_name}-${var.primary_location_short_name}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = 30

  tags = var.tags
}


resource "azurerm_monitor_diagnostic_setting" "vnet_app" {
  name                       = "diag-vnet-app"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  target_resource_id         = azurerm_virtual_network.app_vnet.id
  enabled_log {
    category_group = "allLogs"
  }
  enabled_metric {
    category = "AllMetrics"
  }
}

resource "azurerm_monitor_diagnostic_setting" "vnet_jump" {
  name                       = "diag-vnet-jump"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  target_resource_id         = azurerm_virtual_network.jump_vnet.id
  enabled_log {
    category_group = "allLogs"
  }
  enabled_metric {
    category = "AllMetrics"
  }
}

resource "azurerm_monitor_diagnostic_setting" "azstorage" {
  name                       = "diag-storage"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  target_resource_id         = azurerm_storage_account.main.id
  enabled_metric {
    category = "Transaction"
  }
}

# resource "azurerm_monitor_diagnostic_setting" "key_vault_diag" {
#   name               = "diag-${var.application_name}-${var.environment_name}"
#   target_resource_id = azurerm_key_vault.main.id

#   log_analytics_workspace_id = data.azurerm_log_analytics_workspace.observability.id


#   enabled_log {
#     category_group = "audit"
#   }

#   enabled_log {
#     category_group = "allLogs"
#   }

#   # enabled_log {
#   #   category = "AuditEvent"
#   # }
#   # enabled_log {
#   #   category = "Azure Policy Evaluation Details"
#   # }

#   enabled_metric {
#     category = "AllMetrics"
#   }
# }
