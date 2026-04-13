# S3 Sink Connector — Quick Start

## Overview

The S3 Sink Connector continuously exports data from Kafka topics to Amazon S3, supporting JSON, Avro, Parquet, and raw byte formats. This guide walks you through creating your first S3 Sink Connector on AutoMQ Connect using Terraform.

## Prerequisites

Before creating an S3 Sink Connector, ensure you have:

1. **Running AutoMQ Kafka Instance** — An active Kafka instance in your AutoMQ environment
2. **S3 Bucket** — An Amazon S3 bucket in the same region as your Kafka instance (cross-region is supported but adds latency)
3. **IAM Role** with S3 write permissions:
   - `s3:PutObject`
   - `s3:GetBucketLocation`
   - `s3:ListBucket`
   - `s3:AbortMultipartUpload`
   - `s3:ListMultipartUploadParts`
4. **Kafka User and ACL** — A SASL user with the following ACLs configured:

   | Resource Type | Resource Name | Pattern Type | Operation |
   | --- | --- | --- | --- |
   | TOPIC | your-topic-name (or `*`) | LITERAL | ALL |
   | GROUP | connect (prefix) | PREFIXED | ALL |
   | CLUSTER | kafka-cluster | LITERAL | ALL |
   | TRANSACTIONAL_ID | connect (prefix) | PREFIXED | ALL |

5. **Kubernetes Cluster** — An accessible EKS/K8s cluster registered with CMP for running connector workers

## Terraform Configuration Example

```hcl
# 1. Upload S3 Sink plugin (skip if using a built-in plugin)
resource "automq_connector_plugin" "s3_sink" {
  environment_id  = "env-xxxxx"
  name            = "s3-sink"
  version         = "11.1.0"
  storage_url     = "http://download.automq.com/resource/connector/automq-kafka-connect-s3-11.1.0.zip"
  types           = ["SINK"]
  connector_class = "io.confluent.connect.s3.S3SinkConnector"
}

# 2. Create S3 Sink Connector
resource "automq_connector" "s3_sink" {
  environment_id             = "env-xxxxx"
  name                       = "my-s3-sink"
  plugin_id                  = automq_connector_plugin.s3_sink.id
  connector_class            = "io.confluent.connect.s3.S3SinkConnector"
  plugin_type                = "SINK"
  kubernetes_cluster_id      = "eks-xxxxx"
  kubernetes_namespace       = "default"
  kubernetes_service_account = "default"
  iam_role                   = "arn:aws:iam::123456789012:role/connect-s3-role"
  task_count                 = 3

  capacity = {
    worker_count         = 2
    worker_resource_spec = "TIER2"
  }

  kafka_cluster = {
    kafka_instance_id = automq_kafka_instance.main.id
    security_protocol = {
      security_protocol = "SASL_PLAINTEXT"
      username          = "connect-user"
      password          = var.connect_password
    }
  }

  # Worker config — converter settings (critical!)
  worker_config = {
    "key.converter"                  = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter"                = "org.apache.kafka.connect.json.JsonConverter"
    "value.converter.schemas.enable" = "false"
  }

  # Connector config — S3 Sink parameters
  connector_config = {
    "topics"            = "order-events,user-events"
    "s3.bucket.name"    = "my-data-lake"
    "s3.region"         = "us-east-1"
    "flush.size"        = "1000"
    "rotate.interval.ms" = "600000"
    "storage.class"     = "io.confluent.connect.s3.storage.S3Storage"
    "format.class"      = "io.confluent.connect.s3.format.json.JsonFormat"
    "partitioner.class"  = "io.confluent.connect.storage.partitioner.DefaultPartitioner"
    "topics.dir"        = "kafka-data"
  }

  timeouts = {
    create = "30m"
    delete = "20m"
  }
}
```

## Worker Config (Critical)

The Worker config controls how Kafka Connect serializes/deserializes message keys and values. Incorrect converter settings are the #1 cause of task failures.

| Parameter | Recommended Value | Why |
| --- | --- | --- |
| `key.converter` | `org.apache.kafka.connect.storage.StringConverter` | Most Kafka producers use plain string keys. The default `JsonConverter` will throw `DataException` if the key is not valid JSON. |
| `value.converter` | `org.apache.kafka.connect.json.JsonConverter` | For JSON-formatted message values. Use `AvroConverter` for Avro, `StringConverter` for plain text. |
| `value.converter.schemas.enable` | `false` | Set to `false` if your JSON messages are plain JSON objects. Set to `true` only if messages use the `{schema:{}, payload:{}}` envelope format. |

### Converter Selection Guide

| Key Format | Value Format | key.converter | value.converter | schemas.enable |
| --- | --- | --- | --- | --- |
| String | JSON (no schema) | StringConverter | JsonConverter | false |
| JSON | JSON (no schema) | JsonConverter (schemas.enable=false) | JsonConverter | false |
| String | JSON (with schema) | StringConverter | JsonConverter | true |
| String | Avro | StringConverter | AvroConverter | N/A |
| String | Plain text | StringConverter | StringConverter | N/A |

## Connector Config

### Required Parameters

| Parameter | Description |
| --- | --- |
| `topics` | Comma-separated list of Kafka topics to consume |
| `s3.bucket.name` | Target S3 bucket name |
| `s3.region` | AWS region of the S3 bucket |
| `storage.class` | Fixed: `io.confluent.connect.s3.storage.S3Storage` |
| `format.class` | Output format: `JsonFormat`, `AvroFormat`, `ParquetFormat`, or `ByteArrayFormat` |

### Recommended Parameters

| Parameter | Default | Description |
| --- | --- | --- |
| `flush.size` | (none) | Records per partition before flushing to S3. Recommended: `1000` |
| `rotate.interval.ms` | (none) | Time-based flush interval in ms. Recommended: `600000` (10 min) |
| `partitioner.class` | DefaultPartitioner | S3 directory partitioning strategy |
| `topics.dir` | `topics` | Top-level S3 directory prefix |
| `s3.compression.type` | `none` | Compression: `none` or `gzip` |

## S3 Directory Structure

The connector writes data to S3 using this path pattern:

```
s3://{bucket}/{topics.dir}/{topic}/{partition}/{topic}+{partition}+{startOffset}.{format}
```

Example with `topics.dir=kafka-data`, topic `order-events`, partition 0:

```
s3://my-data-lake/kafka-data/order-events/0/order-events+0+0000000000.json
s3://my-data-lake/kafka-data/order-events/0/order-events+0+0000001000.json
s3://my-data-lake/kafka-data/order-events/1/order-events+1+0000000000.json
```

## Verifying the Connector

After `terraform apply`, verify the connector is running:

1. Check connector state via CMP API:

   ```
   GET /api/v1/connectors/{connectorId}
   ```

   Expect: `state: RUNNING`, `failedTaskCount: 0`

2. Check S3 bucket for output files after the flush interval or flush.size is reached

3. If issues arise, check connector logs:

   ```
   GET /api/v1/connectors/{connectorId}/logs?tailLines=100
   ```

## Troubleshooting

### Task Immediately FAILED After Start

Error log:

```
ERROR WorkerSinkTask Task threw an uncaught and unrecoverable exception
Caused by: org.apache.kafka.connect.errors.DataException:
  Converting byte[] to Kafka Connect data failed due to serialization error
```

**Cause:** The message key or value format does not match the converter configuration.

**Fix:** Update `worker_config` to match your actual message format. Most commonly, change `key.converter` to `StringConverter` (see converter selection guide above).

### No Data Appearing in S3

Check in this order:

1. **Topic has data?** — Verify messages exist in the topic via CMP console
2. **flush.size too large?** — If `flush.size=5000` but only 100 messages exist, data won't flush until the count is reached
3. **rotate.interval.ms set?** — Without time-based rotation, data only flushes when `flush.size` is reached
4. **IAM permissions?** — Ensure the IAM Role has `s3:PutObject` and `s3:ListBucket`
5. **S3 bucket and region correct?** — Verify `s3.bucket.name` and `s3.region` match your actual bucket

### Migrating from Confluent Cloud?

See the [Migration Guide](migration-guide.md) for parameter mapping from Confluent Cloud and MSK Connect.

## Next Steps

- [Configuration Reference](configuration-reference.md) — Full parameter documentation
- [Performance Tuning](performance-tuning.md) — Benchmark data and optimization guidance
- [Troubleshooting](troubleshooting.md) — Comprehensive failure scenario guide
