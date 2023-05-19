#!/usr/bin/env bash
_init_safety_flags_and_DEBUGBASHXTRACE


#####################################################################################################################
# See README.md section "PROCESSING ENGINE peculiarities" about script argument, env-vars, self-contained-environment
#
# Description - this script will:
#   + Validate that ${CollectionName} is lowcase-alphanum-dashes-max18char so it can be used as terraform var.root.collectionName 
#   + Create terraform file variables.auto.tfvars, from content of $JsonOutputFilepath
#   
#####################################################################################################################

#   + Validate that ${CollectionName} is lowcase-alphanum-dashes-max18char so it can be used as terraform var.root.collectionName 
if echo "${CollectionName}" | egrep '^[a-z0-9-]{1,18}$' &>/dev/null
then
  echo "OK - Validation of CollectionName"
else
  echo "NOK - Validation of CollectionName failed"
  exit 1
fi

#   + Create terraform file variables.auto.tfvars, from content of $JsonOutputFilepath
additional_members_of_AADGroup_DataClients="$(cat ${JsonOutputFilepath} | jq -Mec '[.additional_members_of_AADGroup_DataClients.elements[].object_id]')"
  # ["249bbdea-2028-4ed2-8ced-4e784440fe94"]

storageaccounts="{"
for a_storageaccount in $(cat ${JsonOutputFilepath} | jq -Mecr '.storageaccounts.elements[].name')
do 
  storageaccounts+="  \"${a_storageaccount}\" = {}" 
done
storageaccounts+="}"
  # {
  #   "mycollection" = {}
  # } 

cat > ./variables.auto.tfvars <<EOT
root = {
  
  # - Must be lowcase-alphanum-dashes-max30char (must match '^[a-z0-9-]{1,30}$' )
  # catalogName = "default-0-0-0"
  catalogName = "${CatalogName}"

  # - Must be lowcase-alphanum-dashes-max18char (must match '^[a-z0-9-]{1,18}$' )
  # collectionName = "my-collection"
  collectionName = "${CollectionName}"

  # The object_id of user, group or spn
  # additional_members_of_AADGroup_DataClients = ["249bbdea-2028-4ed2-8ced-4e784440fe94"]
  additional_members_of_AADGroup_DataClients = ${additional_members_of_AADGroup_DataClients}


  # - All elements (keys) must be lowercase alphanum no-dashes max18char (must match '^[a-z0-9]{1,18}$' )
  # - Must contain one element with a special name, calculated from the collectionName as explained bellow
  #   Basically remove '-' from collectionName - so if collectionName = "ab-23-c4" then the special name should be "ab23c4"
  #       terraform console
  #       > lower(replace("ab-23-c4", "/[^a-z0-9]/", ""))
  #       "ab23c4"
  #
  #storageaccounts = {
  #  "mycollection" = {}
  #}
  storageaccounts = ${storageaccounts}

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
EOT

