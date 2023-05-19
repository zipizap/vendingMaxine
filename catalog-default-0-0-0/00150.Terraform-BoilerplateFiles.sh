#!/usr/bin/env bash
_init_safety_flags_and_DEBUGBASHXTRACE


#####################################################################################################################
# See README.md section "PROCESSING ENGINE peculiarities" about script argument, env-vars, self-contained-environment
#
# Description - this script will:
#   - Verify required env-vars are set:
#   - Create terraform files: 
#     + provider.azurerm.tf
#     + backend.azurerm.tf
#     + main.tf
#####################################################################################################################

#   - Verify required env-vars are set:
ARM_CLIENT_ID=${ARM_CLIENT_ID?Missing required env-var ARM_CLIENT_ID} 
ARM_CLIENT_SECRET=${ARM_CLIENT_SECRET?Missing required env-var ARM_CLIENT_SECRET} 
ARM_TENANT_ID=${ARM_TENANT_ID?Missing required env-var ARM_TENANT_ID} 
ARM_SUBSCRIPTION_ID=${ARM_SUBSCRIPTION_ID?Missing required env-var ARM_SUBSCRIPTION_ID} 

#   - Create terraform files: 
#     + provider.azurerm.tf
#     + backend.azurerm.tf
#     + main.tf
cat > ./provider.azurerm.tf <<EOT
terraform {
  required_providers {

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.65.0"
    }

    azuread = {
      source = "hashicorp/azuread"
      version = "2.41.0"
    }

  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {
  # Configuration options
}
EOT

  cat > ./backend.azurerm.tf <<EOT
terraform {
  backend "azurerm" {
    resource_group_name  = "vendingmaxine-rg"
    storage_account_name = "tfstate3123"
    container_name       = "tfstates"
    key                  = "${CollectionName}.tfstate"
  }
}
EOT
  
  cat > ./main.tf <<EOT
variable "root" {
  type = object({

    catalogName = string

    collectionName = string

    additional_members_of_AADGroup_DataClients = list(string)

    storageaccounts = map(any)

    keyvaults = map(any)

    spns = map(any)

  })
}
module "mytfmodule" {
  source = "${CatalogDir}/mytfmodule"
  root = var.root
}
EOT
  
