# Only put here variables that require user-input
# Other variables that dont require user-input are better as local-variables; hardcoded values are better inside module code

variable "root" {
  type = object({

    catalogName = string

    collectionName = string

    additional_members_of_AADGroup_DataClients = list(string)

    storageaccounts = map(any)

    keyvaults = map(any)

    spns = map(any)

  })
  description = <<EOT
  Example:
  ```
  root = {
    
    # - Must be lowcase-alphanum-dashes-max30char (must match '^[a-z0-9-]{1,30}$' )
    catalogName = "catalog-default-0-0-0"

    # - Must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' )
    collectionName = "my-collection"

    # The object_id of user, group or spn
    additional_members_of_AADGroup_DataClients = [
      "249bbdea-2028-4ed2-8ced-4e784440fe91",  # AAD user  Alice
      "249bbdea-2028-4ed2-8ced-4e784440fe92"   # AAD group TeamBreeze
      ]

    # - All elements (keys) must be lowercase alphanum no-dashes max18char (must match '^[a-z0-9]{1,18}$' )
    # - Must contain one element with a special name, calculated from the collectionName as explained bellow
    #   Basically remove '-' from collectionName - so if collectionName = "ab-23-c4" then the special name should be "ab23c4"
    #       terraform console
    #       > lower(replace("ab-23-c4", "/[^a-z0-9]/", ""))
    #       "ab23c4"
    #
    storageaccounts = {
      "mycollection" = {}
    }

    # - All elements (keys) must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' )
    # - Must contain one element with a special name, equal to collectionName
    keyvaults = {
      "my-collection" = {}
    }

    # - All elements (keys) must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' )
    # - Must contain one element with a special name, equal to collectionName
    spns = {
      "my-collection" = {}
    }

  }
  ```
EOT

  # Unfortunately condition can only evaluates its own self variable but cannot evaluate other input-vars... 
  #  -> workaround'ed by putting all in 1 var, the var.root
  #
  # NOTE: Its very usefull `terraform console` to troubleshoot validation of tfvar files
  validation {
    error_message = "Failed root.catalogName - Must be lowcase-alphanum-dashes-max30char (must match '^[a-z0-9-]{1,30}$' )."
    condition     = can(regex("^[a-z0-9-]{1,30}$", var.root.catalogName))
  }
  validation {
    error_message = "Failed root.collectionName - Must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' )."
    condition     = length(regexall("^[a-z0-9-]{1,18}$", var.root.collectionName)) == 1
  }
  validation {
    error_message = "Failed root.storageaccounts - All elements (keys) must be lowercase alphanum no-dashes max18char (must match '^[a-z0-9]{1,18}$' )."
    condition     = length( [for k in keys(var.root.storageaccounts) : k    if length(regexall("^[a-z0-9]{1,18}$", k)) == 1] ) == length(keys(var.root.storageaccounts))
  }
  validation {
    error_message = "Failed root.storageaccounts - Must contain one element with a special name, calculated from the collectionName ."
    condition     = contains( keys(var.root.storageaccounts), lower(replace(var.root.collectionName, "/[^a-z0-9]/", "")) )
  }
  validation {
    error_message = "Failed root.keyvaults - All elements (keys) must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' ) ."
    condition     = length( [for k in keys(var.root.keyvaults) : k    if length(regexall("^[a-z0-9-]{1,18}$", k)) == 1] ) == length(keys(var.root.keyvaults))
  }
  validation {
    error_message = "Failed root.keyvaults - Must contain one element with a special name, equal to collectionName ."
    condition     = contains(keys(var.root.keyvaults),var.root.collectionName)
  }
  validation {
    error_message = "Failed root.spns - All elements (keys) must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' ) ."
    condition     = length( [for k in keys(var.root.spns) : k    if length(regexall("^[a-z0-9-]{1,18}$", k)) == 1] ) == length(keys(var.root.spns))
  }
  validation {
    error_message = "Failed root.spns - - Must contain one element with a special name, equal to collectionName ."
    condition     = contains(keys(var.root.spns),var.root.collectionName)
  }
}
