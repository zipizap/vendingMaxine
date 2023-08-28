#!/usr/bin/env bash

show_usage() {
cat <<EOT

$0
  USAGE:
    <CatalogDir>/bin/internal/bash -c "<CatalogDir>/bin/CollectionEdit.Replay.sh <ExistingCollectionEditWorkdir>"

EOT
}

main() {
  [[ "${#}" -eq 0 ]] && show_usage
  local OrigArgs="${@}"
  __dir="$(cd "$(dirname $(readlink -f "${BASH_SOURCE[0]}"))" && pwd)"
  source "${__dir}"/init.selfContainedEnv.source "${@}" "dontSetEnvVars"
  _init_safety_flags_and_DEBUGBASHXTRACE



  ExistingCollectionEditWorkdir="${1?Missing required arg <ExistingCollectionEditWorkdir>}"
  ExistingCollectionEditWorkdir=$(cd ${ExistingCollectionEditWorkdir} && pwd)
    # /a/b/c/collection-name.20230608-234202

  # NewCollectionEditWorkdir gets created in same parent-dir (ParentCollectionEditWorkdir) where ExistingCollectionEditWorkdir exists
  ParentCollectionEditWorkdir="$(cd "$(dirname ${ExistingCollectionEditWorkdir})" && pwd)"
    # /a/b/c
  ExistingCollectionName="$(basename ${ExistingCollectionEditWorkdir} | cut -d. -f1)"
    # collection-name
  NewCollectionEditWorkdir="${ParentCollectionEditWorkdir}/${ExistingCollectionName}.$(date +%Y%m%d-%H%M%S).replay"
    # /a/b/c/collection-name.20230608-235050.replay

  mkdir -p "${NewCollectionEditWorkdir}/CollectionEditFiles"
  [[ -r "${ExistingCollectionEditWorkdir}/CollectionEditFiles/Schema.yaml" ]] && cp "${ExistingCollectionEditWorkdir}/CollectionEditFiles/Schema.yaml" "${NewCollectionEditWorkdir}/CollectionEditFiles/"
  cp "${ExistingCollectionEditWorkdir}/CollectionEditFiles/Schema.json"           "${NewCollectionEditWorkdir}/CollectionEditFiles/"
  cp "${ExistingCollectionEditWorkdir}/CollectionEditFiles/JsonInput.json"        "${NewCollectionEditWorkdir}/CollectionEditFiles/"
  cp "${ExistingCollectionEditWorkdir}/CollectionEditFiles/JsonOutput.orig.json"  "${NewCollectionEditWorkdir}/CollectionEditFiles/"
  cp "${ExistingCollectionEditWorkdir}/CollectionEditFiles/JsonOutput.orig.json"  "${NewCollectionEditWorkdir}/CollectionEditFiles/JsonOutput.json"
  cp "${ExistingCollectionEditWorkdir}/CollectionEditFiles/PeConfig.json"         "${NewCollectionEditWorkdir}/CollectionEditFiles/"

  shw_info "The existing-CollectionEditWorkdir   '${ExistingCollectionEditWorkdir}'"
  shw_info "  will be replayed in new directory  '${NewCollectionEditWorkdir}'"
  exec "${__dir}/internal/bash" -c "${__dir}/CollectionEdit.Launch.sh ${NewCollectionEditWorkdir}"
}
main "${@}"




