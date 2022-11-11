#!/usr/bin/env bash

set -eu

export timeout={{ .Timeout }}

until \
  [ "$(curl -X POST "http://{{ .KibanaHost }}:{{ .KibanaPort }}/api/saved_objects/_import?createNewCopies=true" -H "kbn-xsrf: true" --form file=@{{ .FilePath }} | grep -c 'success')" -ne 0 ]\
; do
    >&2 echo "Waiting for {{ .KibanaHost }}:{{ .KibanaPort }} ...";
  sleep 1;
  timeout=$((timeout - 1));
  if [ $timeout -le 0 ]; then
    echo "Timeout!"
    exit 1;
  fi;
done
