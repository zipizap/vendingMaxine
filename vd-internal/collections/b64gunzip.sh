#set -x
collection="alpha"
#yq_path='.Status.ProcessingEngines[] | select(.Name == "1010.sshkey.engine") | .LatestUpdateData."consumer-selection.next.json"'
yq_path='.Status.Overall.LatestUpdateData."consumer-selection.next.json"'

last_rsf_filepath=$(ls -1rt $collection/RequestStatusFlow.*.yaml | tail -1)
cat $last_rsf_filepath \
| yq "${yq_path}" | base64 -d | gunzip
