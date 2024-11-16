# Sidecar container for telemetry
this is a concept sidecar container that can take metrics for http requests

## what should be done for a production solution
- k8s liveness and readiness probes
- prometheus metrics exporter
- graceful shutdown
- configuration via environment variables
- linter 
- external and internal ports so that metric data does not get into the external network