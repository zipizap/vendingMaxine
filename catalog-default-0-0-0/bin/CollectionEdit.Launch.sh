#!/usr/bin/env bash
show_usage() {
cat <<EOT

$0
  USAGE:
    <CatalogDir>/bin/internal/bash -c "<CatalogDir>/bin/CollectionEdit.Launch.sh <CollectionEditWorkdir>"

EOT
}


main() {
  [[ "${#}" -eq 0 ]] && show_usage
  local OrigArgs="${@}"
  __dir="$(cd "$(dirname $(readlink -f "${BASH_SOURCE[0]}"))" && pwd)"
  source "${__dir}"/init.selfContainedEnv.source "${@}"
  _init_safety_flags_and_DEBUGBASHXTRACE

  shw_info "Run ProcEngines in sequence, inside CollectionEditWorkdir '${CollectionEditWorkdir}'"
  for a_Pe in $(find ${CatalogDir} -type f -executable -follow -maxdepth 1 | sort )
  do
    # a_Pe:             <CatalogDir>/00101.resource-group.Procesor.sh
    # a_PeName:         00101.resource-group.Procesor.sh
    # a_PeTracklogsdir: <PeTracklogsParentdir>/00101.resource-group.Procesor.sh
    local a_PeName="$(basename ${a_Pe})"
    local a_PeTracklogsdir="${PeTracklogsParentdir}/${a_PeName}"
    mkdir -p "${a_PeTracklogsdir}"
    # pre TrackLogs files
    [[ -r "${CollectionEditWorkdir}/CollectionEditFiles/Schema.yaml" ]] && cp "${CollectionEditWorkdir}/CollectionEditFiles/Schema.yaml"  "${a_PeTracklogsdir}"
    cp "${CollectionEditWorkdir}/CollectionEditFiles/Schema.json"      "${a_PeTracklogsdir}"
    cp "${CollectionEditWorkdir}/CollectionEditFiles/JsonInput.json"   "${a_PeTracklogsdir}"
    cp "${CollectionEditWorkdir}/CollectionEditFiles/JsonOutput.json"  "${a_PeTracklogsdir}"/JsonOutput.beginning.json
    
    # Run inside PWD=CollectionEditWorkdir
    shw_info "++ '$a_PeName' - Run starting (a_PeTracklogsdir '${a_PeTracklogsdir}')"
    cd "${CollectionEditWorkdir}"
    
    # Call a_Pe PE_CONFIG_JSON_FILE
    set +o errexit
    "${a_Pe}" "${PE_CONFIG_JSON_FILE}" 2>&1 | tee "${a_PeTracklogsdir}"/StdoutStderr.txt
    local a_PeExitCode="${PIPESTATUS[0]}"
      # 0 or 1 or ...
    shw_info "-- '$a_PeName' - Run finished, exit-code '${a_PeExitCode}' (a_PeTracklogsdir '${a_PeTracklogsdir}')"
    set -o errexit

    # post TrackLogs files
    echo "${a_PeExitCode}" > "${a_PeTracklogsdir}"/ExitCode.txt
    cp "${CollectionEditWorkdir}/CollectionEditFiles/JsonOutput.json"  "${a_PeTracklogsdir}"/JsonOutput.json

    if [[ ${a_PeExitCode} -ne 0 ]]
    then
      # a_Pe had exit-code != 0
      # So lets  exit with that exit-code
      exit ${a_PeExitCode}
    fi

  done

  shw_info "== Finished successfully =="
}
main "${@}"