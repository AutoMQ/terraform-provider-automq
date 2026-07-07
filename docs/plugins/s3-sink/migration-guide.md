# S3 Sink Connector — Migration Guide

## Migrating from Confluent Cloud

### Parameter Mapping

| # | Confluent Cloud Parameter | AutoMQ Equivalent | Notes |
| --- | --- | --- | --- |
| 1 | `connector.class = S3_SINK` | `connector_class = io.confluent.connect.s3.S3SinkConnector` | Confluent uses short name, AutoMQ uses fully-qualified class name |
| 2 | `output.data.format = JSON` | `format.class = io.confluent.connect.s3.format.json.JsonFormat` | See format mapping table below |
| 3 | `output.data.format = AVRO` | `format.class = io.confluent.connect.s3.format.avro.AvroFormat` | |
| 4 | `output.data.format = PARQUET` | `format.class = io.confluent.connect.s3.format.parquet.ParquetFormat` | |
| 5 | `output.data.format = BYTES` | `format.class = io.confluent.connect.s3.format.bytearray.ByteArrayFormat` | |
| 6 | `input.data.format` | Not needed | Confluent-only: controls deserialization on Confluent's managed platform |
| 7 | `kafka.auth.mode` | Not needed | Auth configured via `kafka_cluster.security_protocol` in Terraform |
| 8 | `kafka.api.key` | Not needed | Replace with SASL username in `kafka_cluster.security_protocol.username` |
| 9 | `kafka.api.secret` | Not needed | Replace with SASL password in `kafka_cluster.security_protocol.password` |
| 10 | `kafka.service.account.id` | Not needed | Confluent-only service account concept |
| 11 | `tasks.max` | `task_count` | Terraform resource attribute (not in connector_config) |
| 12 | `topics` | `connector_config["topics"]` | Direct copy |
| 13 | `s3.bucket.name` | `connector_config["s3.bucket.name"]` | Direct copy |
| 14 | `s3.region` | `connector_config["s3.region"]` | Direct copy |
| 15 | `flush.size` | `connector_config["flush.size"]` | Direct copy |
| 16 | `rotate.interval.ms` | `connector_config["rotate.interval.ms"]` | Direct copy |
| 17 | `s3.compression.type` | `connector_config["s3.compression.type"]` | Direct copy |
| 18 | `partitioner.class` | `connector_config["partitioner.class"]` | Direct copy (if set) |
| 19 | `topics.dir` | `connector_config["topics.dir"]` | Direct copy (if set) |

### Confluent-Only Parameters (Do Not Migrate)

These parameters are specific to Confluent Cloud's managed platform and have no equivalent in AutoMQ:

| Parameter | Why Not Needed |
| --- | --- |
| `kafka.auth.mode` | Confluent uses API Key auth. AutoMQ uses SASL configured in `kafka_cluster.security_protocol`. |
| `kafka.api.key` | Replaced by SASL username. |
| `kafka.api.secret` | Replaced by SASL password. |
| `kafka.service.account.id` | Confluent service account concept. AutoMQ uses Kafka users with ACLs. |
| `input.data.format` | Confluent-specific deserialization control. AutoMQ uses standard Kafka Connect converters in `worker_config`. |
| `output.data.format` | Mapped to `format.class` (see mapping table above). |

### AutoMQ-Specific Parameters (Not in Confluent)

Confluent Cloud is fully managed, so these parameters don't exist there but are required in AutoMQ:

| Parameter | Description |
| --- | --- |
| `storage.class` | Fixed: `io.confluent.connect.s3.storage.S3Storage`. Confluent sets this automatically. |
| `key.converter` (worker_config) | Worker-level key converter. Confluent handles this internally. |
| `value.converter` (worker_config) | Worker-level value converter. Confluent handles this internally. |
| `value.converter.schemas.enable` (worker_config) | Schema wrapper control. Confluent handles this internally. |
| `capacity.worker_count` | Number of worker pods. Confluent auto-scales. |
| `capacity.worker_resource_spec` | Worker resource tier (TIER1–TIER4). Confluent abstracts compute. |
| `kubernetes_cluster_id` | K8s cluster for worker deployment. Confluent is serverless. |
| `kubernetes_namespace` | K8s namespace. Confluent is serverless. |
| `iam_role` | IAM Role for S3 access. Confluent uses its own credential management. |
| `kafka_cluster.security_protocol` | Kafka authentication config. Confluent uses API Keys internally. |

### Migration Steps (Confluent Cloud → AutoMQ)

1. **Export Confluent connector config** — From Confluent Cloud console or API, export the connector's `config_nonsensitive` and `config_sensitive` parameters
2. **Map connector.class** — Change `S3_SINK` to `io.confluent.connect.s3.S3SinkConnector`
3. **Map output.data.format to format.class** — Use the format mapping table above (e.g., `JSON` → `io.confluent.connect.s3.format.json.JsonFormat`)
4. **Remove Confluent-only parameters** — Delete `kafka.auth.mode`, `kafka.api.key`, `kafka.api.secret`, `kafka.service.account.id`, `input.data.format`, `output.data.format`
5. **Add storage.class** — Set `storage.class = io.confluent.connect.s3.storage.S3Storage`
6. **Configure worker converters** — Set `key.converter`, `value.converter`, and `value.converter.schemas.enable` in `worker_config` based on your message format
7. **Create SASL user** — Create a Kafka user with SASL credentials to replace Confluent API Keys
8. **Configure ACLs** — Grant TOPIC, GROUP, CLUSTER, and TRANSACTIONAL_ID ACLs to the SASL user
9. **Choose capacity** — Select worker tier and count based on your throughput needs (see [Performance Tuning Guide](performance-tuning.md))
10. **Configure K8s deployment** — Set `kubernetes_cluster_id`, `kubernetes_namespace`, `kubernetes_service_account`, and `iam_role`
11. **Apply Terraform** — Run `terraform plan` and `terraform apply` to create the connector
12. **Verify** — Check connector state is RUNNING and data appears in S3

## Migrating from MSK Connect

### Parameter Mapping

| # | MSK Connect Parameter | AutoMQ Equivalent | Notes |
| --- | --- | --- | --- |
| 1 | `connector.class` | `connector_class` | Direct copy (MSK uses fully-qualified class name) |
| 2 | `tasks.max` | `task_count` | Direct copy |
| 3 | `topics` | `connector_config["topics"]` | Direct copy |
| 4 | `s3.bucket.name` | `connector_config["s3.bucket.name"]` | Direct copy |
| 5 | `s3.region` | `connector_config["s3.region"]` | Direct copy |
| 6 | `flush.size` | `connector_config["flush.size"]` | Direct copy |
| 7 | Other connector config | `connector_config[...]` | Most S3 Sink params copy directly |
| 8 | `capacity.mcuCount` | `capacity.worker_resource_spec` | MCU → Tier mapping (see table below) |
| 9 | `capacity.workerCount` | `capacity.worker_count` | Direct copy |
| 10 | `plugins[].customPlugin.arn` | Re-upload via `automq_connector_plugin` | MSK plugin ARNs don't transfer |
| 11 | `serviceExecutionRoleArn` | `iam_role` | Direct copy of the IAM Role ARN |
| 12 | `workerConfiguration` | `worker_config` | Extract key-value pairs from MSK worker configuration resource |

### MCU → Tier Mapping

| MSK MCU Count | AutoMQ Tier | CPU | Memory |
| --- | --- | --- | --- |
| 1 MCU | TIER1 | 0.5 | 1 GiB |
| 2 MCU | TIER2 | 1 | 2 GiB |
| 4 MCU | TIER3 | 2 | 4 GiB |
| 8 MCU | TIER4 | 4 | 8 GiB |

### Migration Steps (MSK Connect → AutoMQ)

1. **Export MSK connector config** — From AWS Console or `aws kafkaconnect describe-connector`, export the `connectorConfiguration` map
2. **Upload plugin** — Create an `automq_connector_plugin` resource with the same S3 Sink plugin JAR (download from Confluent Hub or use AutoMQ's hosted URL)
3. **Split configuration** — Separate MSK's flat `connectorConfiguration` into `connector_config` (S3 Sink params) and `worker_config` (converter params)
4. **Map MCU to Tier** — Convert MSK's `capacity.mcuCount` to AutoMQ's `worker_resource_spec` using the mapping table above
5. **Configure Kafka authentication** — MSK may use IAM auth; AutoMQ uses SASL. Create a Kafka user and configure `kafka_cluster.security_protocol`
6. **Configure K8s deployment** — Set `kubernetes_cluster_id`, `kubernetes_namespace`, `kubernetes_service_account`
7. **Map IAM Role** — Copy `serviceExecutionRoleArn` to `iam_role` (ensure the role has S3 write permissions)
8. **Apply Terraform** — Run `terraform plan` and `terraform apply`
9. **Verify** — Check connector state is RUNNING and data appears in S3 with the expected directory structure
