#!/usr/bin/env bash
_init_safety_flags_and_DEBUGBASHXTRACE

#####################################################################################################################
# See README.md section "PROCESSING ENGINE peculiarities" about script argument, env-vars, self-contained-environment
#
# Description - this script will:
#  + Call terraform init, plan, apply
#####################################################################################################################

#  + Call terraform init, plan, apply
# terraform init
terraform init -input=false  

# terraform validate
terraform validate 

# terraform plan+apply
terraform plan -lock=false
#terraform apply -auto-approve 

# # terraform destroy would be:
# terraform destroy \
#    -input=false -auto-approve

# clean terraform tmp files that are quite big, to relieve db and replay file size
rm -rf ./.terraform* &>/dev/null