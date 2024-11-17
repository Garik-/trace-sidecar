# Sidecar container for telemetry
this is a concept sidecar container that can take metrics for http requests

## what should be done for a production solution
- ~~k8s liveness and readiness probes~~
- ~~prometheus metrics exporter~~
- ~~graceful shutdown~~
- ~~configuration via environment variables~~
- ~~linter~~
- ~~external and internal ports so that metric data does not get into the external network~~
- ~~route metrics~~
- load testing
- dockerfile
- local deploy

## server lifecycle
![server lifecycle](https://ieftimov.com/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/request-lifecycle-timeouts.png "server lifecycle")