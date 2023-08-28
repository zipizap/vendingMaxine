#---------------------------------------------------
# Helpfull objects
data "azurerm_client_config" "current" {}

resource "random_string" "random6charsuffix" {
  length  = 6
  special = false
  upper = false
}


#---------------------------------------------------
# locals
locals {
  tags = {
    "vendingmaxine" = "${var.root.catalogName}.${var.root.collectionName}"
    "collection" = var.root.collectionName
    #cost-center
    #team
  }
  randomsuffix = random_string.random6charsuffix.result
  location = "West Europe"
  base_sta_inputvarkey = lower(replace(var.root.collectionName, "/[^a-z0-9]/", ""))
  base_kvt_inputvarkey = var.root.collectionName
  base_spn_inputvarkey = var.root.collectionName
}

#------------------------------------------------
# AAD: Group "mycollection_DataClients" (base_AADGroup_dataClients)
resource "azuread_group" "dataClients" {
  display_name = "${var.root.collectionName}_DataClients"
  owners           = [data.azurerm_client_config.current.object_id]
  security_enabled = true
}
locals {
   base_AADGroup_dataClients = azuread_group.dataClients
}
resource "azuread_group_member" "additional_members_of_AADGroup_DataClients" {
  for_each = toset(var.root.additional_members_of_AADGroup_DataClients)
  group_object_id  = azuread_group.dataClients.id
  member_object_id = each.key
}



#------------------------------------------------
# ResourceGroup + StorageAccount + Keyvault
resource "azurerm_resource_group" "rg" {
  name     = var.root.collectionName
  location = local.location
  tags = local.tags
}

resource "azurerm_storage_account" "sta" {
  for_each = var.root.storageaccounts
  name                     = "${each.key}${local.randomsuffix}"
  location                 = azurerm_resource_group.rg.location
  resource_group_name      = azurerm_resource_group.rg.name
  account_tier             = "Standard"
  account_replication_type = "LRS"
  tags = local.tags
}

resource "azurerm_key_vault" "kvt" {
  for_each = var.root.keyvaults
  name                     = "${each.key}${local.randomsuffix}"
  location                 = azurerm_resource_group.rg.location
  resource_group_name      = azurerm_resource_group.rg.name
  tenant_id                = data.azurerm_client_config.current.tenant_id

  sku_name = "standard"
  enable_rbac_authorization = true
  purge_protection_enabled    = false

  tags = local.tags
}
locals {
  base_sta = azurerm_storage_account.sta[local.base_sta_inputvarkey]
  base_kvt = azurerm_key_vault.kvt[local.base_kvt_inputvarkey]
}
# save tennant_id and subscription_id into base_kvt secrets
resource "azurerm_key_vault_secret" "tenant_id" {
  for_each = var.root.spns
  name         = "TENANTID"
  value        = data.azurerm_client_config.current.tenant_id
  key_vault_id = local.base_kvt.id
}
resource "azurerm_key_vault_secret" "subscription_id" {
  for_each = var.root.spns
  name         = "SUBSCRIPTIONID"
  value        = data.azurerm_client_config.current.subscription_id
  key_vault_id = local.base_kvt.id
}



#------------------------------------------------
# AppReg + Spn + SpnPasswd + base_kvt secret + base_AADGroup_dataClients membership
resource "azuread_application" "app" {
  for_each = var.root.spns
  display_name = each.key
  owners       = [data.azurerm_client_config.current.object_id]
}
resource "azuread_service_principal" "spn" {
  for_each = var.root.spns
  application_id = azuread_application.app[each.key].application_id
  owners       = [data.azurerm_client_config.current.object_id]
}
resource "time_rotating" "spnpasswd" {
  for_each = var.root.spns
  rotation_days = 30
}
resource "azuread_service_principal_password" "spnpasswd" {
  for_each = var.root.spns
  service_principal_id = azuread_service_principal.spn[each.key].object_id
  rotate_when_changed = {
    rotation = time_rotating.spnpasswd[each.key].id
  }
}
locals {
  base_app = azuread_application.app[local.base_spn_inputvarkey]
  base_spn = azuread_service_principal.spn[local.base_spn_inputvarkey]
}
# save spn info into base_kvt secrets
resource "azurerm_key_vault_secret" "spn_client_secret" {
  for_each = var.root.spns
  name         = "${azuread_service_principal.spn[each.key].display_name}-SPN-CLIENTSECRET"
  value        = azuread_service_principal_password.spnpasswd[each.key].value
  key_vault_id = local.base_kvt.id
}
resource "azurerm_key_vault_secret" "spn_client_id" {
  for_each = var.root.spns
  name         = "${azuread_service_principal.spn[each.key].display_name}-SPN-CLIENTID"
  value        = azuread_service_principal.spn[each.key].application_id
  key_vault_id = local.base_kvt.id
}
# add spn as member of base_AADGroup_dataClients
resource "azuread_group_member" "spn" {
  for_each = var.root.spns
  group_object_id  = azuread_group.dataClients.id
  member_object_id = azuread_service_principal.spn[each.key].id
}


#------------------------------------------------
# RBAC Role assignments
#
# - roleassignments of base_AADGroup_dataClients over rg, affecting storageaccounts: 
#     . Storage Blob Data Owner
#     . Storage File Data Privileged Contributor
# - roleassignments of base_AADGroup_dataClients over rg, affecting keyvaults: 
#     . Key Vault Administrator
locals {
  tmpRolesToAssign = [
    "Storage Blob Data Owner",
    "Storage File Data Privileged Contributor",
    "Key Vault Administrator",
  ]
}
resource "azurerm_role_assignment" "dataClients_on_rg" {
  for_each = toset(local.tmpRolesToAssign)
  scope                = azurerm_resource_group.rg.id
  principal_id         = azuread_group.dataClients.object_id
  role_definition_name = each.key
}

