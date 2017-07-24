#!/bin/sh

docker run -l io.kubernetes.pod.namespace=$1 \
           -l io.kubernetes.pod.name=$2 \
           -l io.kubernetes.container.name=$3 \
           busybox /bin/sh -c 'i=0; while true; do echo "$i: $(date)"; i=$((i+1)); sleep 1; done'
