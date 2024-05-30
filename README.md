Firebolt OpenTelemetry Exporter
-------------------------------
Firebolt OpenTelemetry Exporter is designed to simplify exporting metrics of running engines in [Firebolt](https://www.firebolt.io/) organization.

The exporter is based on  [OTLP](https://opentelemetry.io/docs/specs/otel/protocol/) and offers a standardized way of pushing metrics to the 
[OpenTelemetry collector](https://opentelemetry.io/docs/collector/) of your choice.

In order to use the exporter, you need to have an organization registered with Firebolt since exporter will use
your Firebolt organization account to access accounts/engines and collect metrics.

Exporter is provided as a docker image that can be run in your infrastructure.

Pre-requisites before use
-------------------------
1. Set up an OpenTelemetry collector of your choice. Most of the cloud monitoring solutions, such as Datadog, Coralogix, 
Prometheus etc., support OTLP standard and offer bundled OpenTelemetry collectors.
2. Create a Service Account in your Firebolt Organization, and grant it with permissions in accounts which you are
going to monitor. 

    NOTE: exporter will query `information_schema.engine_metrics_history` and `information_schema.engine_query_history` views,
so make sure that permission model allows Service Account use these views.

Find more details on how to create a Service Account in [Firebolt documentation](https://docs.firebolt.io/godocs/Guides/managing-your-organization/service-accounts.html).
3. Run `otel-exporter` docker container.

Example use
-----------
Here's the example how to start `otel-exporter` as a docker container on localhost, assuming there's a collector listening 
on 127.0.0.1:4317 via GRPC:
```shell
docker run --name firebolt-otel-exporter \
  -e FIREBOLT_OTEL_EXPORTER_CLIENT_ID=<service_account_client_id> \
  -e FIREBOLT_OTEL_EXPORTER_CLIENT_SECRET=<service_account_client_secret> \
  -e FIREBOLT_OTEL_EXPORTER_ACCOUNTS=my-account1,my-account2 \
  -e FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS=127.0.0.1:4317 \
  -e FIREBOLT_OTEL_EXPORTER_LOG_LEVEL=debug \
  --network="host" \
ghcr.io/firebolt-db/otel-exporter:v0.0.1
```

Metrics
-------
The exporter's structure of meters and metrics is described below. See [OTLP metrics API](https://opentelemetry.io/docs/specs/otel/metrics/api/) for reference. 

#### Meter name: `firebolt.engine.runtime`

| Metric                             | Type              | Description                                                                                |
|------------------------------------|-------------------|--------------------------------------------------------------------------------------------|
| firebolt.engine.cpu.utilization    | Float64Gauge      | Current CPU utilization (percentage)                                                       |
| firebolt.engine.memory.utilization | Float64Gauge      | Current Memory used (percentage)                                                           |
| firebolt.engine.disk.utilization   | Float64Gauge      | Currently used disk space which encompasses space used for cache and spilling (percentage) |
| firebolt.engine.cache.utilization  | Float64Gauge      | Current SSD cache hit ratio (percentage)                                                   |
| firebolt.engine.disk.spilled       | In64UpDownCounter | Amount of spilled data to disk (byte)                                                      |
| firebolt.engine.running.queries    | Int64Gauge        | Number of running queries                                                                  |
| firebolt.engine.suspended.queries  | Int64Gauge        | Number of suspended queries                                                                |


#### Meter name: `firebolt.engine.query_history`

| Metric                       | Type             | Description                                                     |
|------------------------------|------------------|-----------------------------------------------------------------|
| firebolt.query.duration      | Float64Histogram | Duration of query execution (second)                            |
| firebolt.query.scanned.rows  | Int64Counter     | The total number of rows scanned                                |
| firebolt.query.scanned.byte  | Int64Counter     | The total number of bytes scanned (both from cache and storage) |
| firebolt.query.insert.rows   | Int64Counter     | The total number of rows written                                |
| firebolt.query.insert.byte   | Int64Counter     | The total number of bytes written (both to cache and storage)   |
| firebolt.query.returned.rows | Int64Counter     | The total number of rows returned from the query                |
| firebolt.query.returned.byte | Int64Counter     | The total number of bytes returned from the query               |
| firebolt.query.spilled.byte  | Int64Counter     | The total number of bytes spilled (uncompressed)                |
| firebolt.query.queue.time    | Float64Counter   | Time the query spent in queue                                   |

#### Meter name: `firebolt.exporter`

| Metric                      | Type            | Description                                     |
|-----------------------------|-----------------|-------------------------------------------------|
| firebolt.exporter.duration  | Float64Counter  | Duration of collection routine of the exporter  |

Configuration reference
-----------------------
All the configuration variables are passed as environment variables. Variables have prefix `FIREBOLT_OTEL_EXPORTER_*`.

| Parameter                    | Required                       | Description                                                                                                                                                   | Default value |
|------------------------------|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| CLIENT_ID                    | Yes                            | Client ID derived from the Service Account                                                                                                                    |               |
| CLIENT_SECRET                | Yes                            | Client Secret derived from the Service Account                                                                                                                |               |
| ACCOUNTS                     | Yes                            | List of accounts to monitor. The Service Account needs to have access to all theseaccounts to be able to fetch metrics data. At least one account is required |               |
| COLLECT_INTERVAL             | No                             | Defines how often metrics will be collected. Ninimal allowed value is 15s                                                                                     | `30s`         |
| LOG_FORMAT                   | No                             | Log format, either `json` or `text`                                                                                                                           | `json`        |
| LOG_LEVEL                    | No                             | Log level, one of `debug`, `info`, `error`                                                                                                                    | `info`        |
| GRPC_ADDRESS                 | Yes, if GRPC collector is used | GRPC address of collector, where metrics will be pushed, for example `127.0.0.1:4317`                                                                         |               |
| HTTP_ADDRESS                 | Yes, if HTTP collector is used | HTTP address of collector, where metrics will be pushed, for example `127.0.0.1:4318`                                                                         |               |
| GRPC_OAUTH_CLIENT_ID         | No                             | OAuth2 client id                                                                                                                                              |               |
| GRPC_OAUTH_CLIENT_SECRET     | No                             | OAuth2 client secret                                                                                                                                          |               |
| GRPC_OAUTH_TOKEN_URL         | No                             | OAuth2 resource server's token endpoint URL                                                                                                                   |               |
| SYSTEM_CERT_POOL             | No                             | Enables TLS security based on operating system certificate pool (`true` or `false`)                                                                           | `false`       |
| HTTP_TLS_X509_CERT_PEM_BLOCK | No                             | Specifies TLS certificate PEM in case HTTP mTLS authentication is used                                                                                        |               |
| HTTP_TLS_X509_KEY_PEM_BLOCK  | No                             | Specifies TLS key PEM in case HTTP mTLS authentication is used                                                                                                |               |

**NOTE:** Either `FIREBOLT_OTEL_EXPORTER_GRPC_ADDRESS` or `FIREBOLT_OTEL_EXPORTER_HTTP_ADDRESS` must be provided.
