yq '.Status.ProcessingEngines[] | select(.Name == "1010.sshkey.engine") | .LatestUpdateData."consumer-selection.next.json"'  RequestStatusFlow.20220705041109.48.yaml | base64 -d | gunzip
