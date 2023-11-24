#!/usr/bin/env bash
# Paulo Aleixo Campos
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
function shw_info { echo -e '\033[1;34m'"$1"'\033[0m'; }
function error { echo "ERROR in ${1}"; exit 99; }
trap 'error $LINENO' ERR
PS4='████████████████████████${BASH_SOURCE}@${FUNCNAME[0]:-}[${LINENO}]>  '
set -o errexit
set -o pipefail
set -o nounset
#set -o xtrace


cd "${__dir}"
export GO111MODULES=on
[[ -r local_env.source ]] && source local_env.source
./swagger_update.sh
#exec go run . ${@} 2>&1 


exec go run . ${@} 2>&1 \
| hl -g ':200}' -R ':[45][0-9][0-9]}' -B '\[PerID[^\]]+\]'  


# clear; rm sqlite.db ; ./go_run.sh
# clear; rm sqlite.db ; VD_LOGS_LOGLEVEL="DEBUG" ; ./go_run.sh
 

