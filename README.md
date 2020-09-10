
# Prometheus Multi-Tenant Proxy

Two micro-services:

1) Sits between the metrics provider and Prometheus, enriching the metrics with a configurable `namespace`.
2) Sits between Prometheus and Grafana to inject the `namespace` into the query based on a claim in a JWT token.

Example:

Kong --> `enrich-proxy` --> Prometheus <-- `query-proxy` <-- Grafana <-- Keycloak
