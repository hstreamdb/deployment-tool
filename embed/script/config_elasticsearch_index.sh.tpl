#!/usr/bin/env bash

set -eu

max_attempts=60
wait_seconds=1
ES_URL="http://{{.Host }}:{{ .Port }}"

for ((i=1; i<=max_attempts; i++)); do
    # Check if Elasticsearch is up by querying its root endpoint
    if curl -f -s "${ES_URL}" > /dev/null; then
        break
    else
        sleep ${wait_seconds}
    fi

    if [ $i -eq ${max_attempts} ]; then
        echo "Failed to connect to Elasticsearch after ${max_attempts} attempts."
        exit 1
    fi
done

# Define an index lifecycle policy, stay 1 day in hot phase, 6 day in warm phase then delete
curl -X PUT "${ES_URL}/_ilm/policy/log_policy?pretty" -H 'Content-Type: application/json' -d'
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {}
      },
      "warm": {
        "min_age": "1d",
        "actions": {}
      },
      "delete": {
        "min_age": "6d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
'

# Define an index template for hstream-log-* indices. For now, we only run es in single-node mode, so
# the number of replicas is set to 0.
curl -X PUT "${ES_URL}/_index_template/hstream_log_template?pretty" -H 'Content-Type: application/json' -d'
{
  "index_patterns": ["hstream-log-*"],
  "template": {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "index.max_result_window": 1000000,
      "index.lifecycle.name": "log_policy"
    }
  }
}
'
