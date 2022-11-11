#!/usr/bin/env bash

set -eu

cnt=$(({{.Shard}}-1))

for i in $(seq 0 ${cnt})
do
  shardCnt=$((i%{{.Disk}}))
  mkdir -p /mnt/data${shardCnt}/shard${i} && ln -s /mnt/data${shardCnt}/shard${i} {{.DataDir}}/shard${i}
done