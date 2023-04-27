#!/usr/bin/env bash
counter=1
for a_arg in "${@}"; do
  echo ">> arg $counter:  '$a_arg'"
  counter=$(( counter + 1 ))
done
