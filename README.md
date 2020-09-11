
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

### Metrics Proxy

```
export MYAPP_PORT=9092
export MYAPP_METRICSURL=https://kong-admin:8001
export MYAPP_LABELMAPPATH=sample-map.yml

bin/metrics-proxy
```

### Query Proxy

```
export MYAPP_PORT=9091
export MYAPP_PROMETHEUSURL=http://prometheus-server:9090
export MYAPP_NAMESPACELABEL=team
export MYAPP_NAMESPACECLAIM=team
export MYAPP_JWKSURL=https://auth.org/auth/realms/myrealm/protocol/openid-connect/certs
export MYAPP_VERIFYTOKEN=false

bin/query-proxy
```

```
export TOK=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE5MTYyMzkwMjIsInRlYW0iOiJhYmMifQ.bdmbECR2RdUCRxgSpY8hxQ0aRYlKyvHZxRfoinLUeA0

export QUERY="sum(rate(kong_http_status%7Binstance%3D~%22.*%22%7D%5B1m%5D))"

curl -v "http://localhost:9091/api/v1/query_range?step=14&start=1599844738.000&end=1599848338.000&query=$QUERY" -H "Authorization: Bearer $TOK"

```

## Docker

### Metrics Proxy

```
docker build --tag metrics-proxy.local -f metrics-proxy/Dockerfile metrics-proxy

docker run --rm --name metrics-proxy \
  -p 9091:9091 \
  -e MYAPP_PORT=9091 \
  -e MYAPP_METRICSURL=https://metrics_providers/metrics \
  -e MYAPP_LABELMAPPATH=/var/sample-map.yml \
  -v `pwd`/metrics-proxy/sample-map.yml:/var/sample-map.yml \
  metrics-proxy.local
```

### Query Proxy

docker build --tag query-proxy.local -f query-proxy/Dockerfile query-proxy

docker run --rm --name query-proxy \
  -p 9092:9092 \
  -e MYAPP_PORT=9092 \
  -e MYAPP_NAMESPACELABEL=team \
  -e MYAPP_NAMESPACECLAIM=team \
  -e MYAPP_PROMETHEUSURL=http://prometheus-server:9090 \
  -e MYAPP_JWKSURL=https://auth.org/auth/realms/myrealm/protocol/openid-connect/certs \
  query-proxy.local
```
