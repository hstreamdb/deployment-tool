#!/usr/bin/env bash

set -eu

export timeout={{.Timeout}}

until (echo -n > /dev/tcp/{{.Host}}/{{.Port}}); do
  >&2 echo "Waiting for {{.Host}}:{{.Port}} ...";
  sleep 1;
  timeout=$((timeout - 1));
  if [ $timeout -le 0 ]; then
    echo "Timeout!"
    exit 1;
  fi;
done

sleep 4
