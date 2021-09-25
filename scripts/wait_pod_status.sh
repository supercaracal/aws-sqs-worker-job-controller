#!/bin/bash

set -euo pipefail

readonly NAME=$1
readonly STATUS=$2

while :
do
  sts=($(\
    kubectl --context=kind-kind get pods -o json | \
    jq -r ".items[] | select(.metadata.name | contains(\"${NAME}\")) | .status.phase" | \
    sort | uniq))

  if [[ ${#sts[@]} -eq 1 && $sts = $STATUS ]]; then
    break
  fi

  echo "Waiting for ${NAME} to be ${STATUS}..."
  sleep 3
done
