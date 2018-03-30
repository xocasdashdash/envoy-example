#!/usr/bin/env bash
set -e
#Check configuration
cat /etc/envoy/front-envoy.yaml
/usr/local/bin/envoy -c /etc/envoy/front-envoy.yaml --v2-config-only --mode validate --service-cluster front-proxy --restart-epoch $RESTART_EPOCH
exec /usr/local/bin/envoy -c /etc/envoy/front-envoy.yaml --service-cluster front-proxy --restart-epoch $RESTART_EPOCH
