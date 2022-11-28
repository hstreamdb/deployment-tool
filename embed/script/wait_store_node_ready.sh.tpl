#!/usr/bin/env bash

set -eu

export timeout={{.Timeout}}

until (echo -n > /dev/tcp/{{.Host}}/{{.AdminApiPort}}); do
  >&2 echo "Waiting for {{.Host}}:{{.AdminApiPort}} ...";
  sleep 1;
  timeout=$((timeout - 1));
  if [ $timeout -le 0 ]; then
    echo "Timeout!"
    exit 1;
  fi;
done

sleep 4
