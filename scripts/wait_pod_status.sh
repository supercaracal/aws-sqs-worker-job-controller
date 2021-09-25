#!/bin/bash

set -euo pipefail

readonly NAME=$1
readonly STATUS=$2

while :
do
  sts=$(kubectl --context=kind-kind get pods -o json | jq -r ".items[] | select(.metadata.name | contains(\"${NAME}\")) | .status.phase")
  if [[ $sts = $STATUS ]]; then
    break
  fi
  echo "Waiting for ${NAME} to be ready..."
  sleep 3
done
