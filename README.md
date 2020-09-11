
# Prometheus Multi-Tenant Proxy

Two micro-services:

1) Sits between the metrics provider and Prometheus, enriching the metrics with a configurable `namespace`.
2) Sits between Prometheus and Grafana to inject the `namespace` into the query based on a claim in a JWT token.

Example:

Kong --> `metrics-proxy` --> Prometheus <-- `query-proxy` <-- Grafana <-- Keycloak


## Building

```
cd metrics-proxy
make go-build
bin/metrics-proxy

cd query-proxy
make go-build
bin/query-proxy
```

## Running

```
export MYAPP_PORT=9092
export MYAPP_METRICSURL=https://kong-admin:8001
export MYAPP_LABELMAPPATH=sample-map.yml

bin/metrics-proxy
