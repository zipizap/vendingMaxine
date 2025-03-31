#!/usr/bin/env bash

# USAGE:
#   cd mydir
#   go_boilerplate_here.sh 

cp -v /t/Seafile/SharedUnsec/Refs/go/cliapp_boilerplate/{*,.*} . || true
cat <<EOT
NOW:
 - code main.go
 - ./go_mod_and_run.sh
EOT
code . 


