# S3 Sink Connector — Configuration Reference

This document covers all configuration parameters for the S3 Sink Connector (io.confluent.connect.s3.S3SinkConnector v11.1.0), extracted from the S3SinkConnectorConfig ConfigDef (42 parameters total, 4 sensitive).

## Worker Config

These parameters are set in `worker_config` and affect the entire Connect Worker.

| Parameter | Type | Recommended Value | Description |
| --- | --- | --- | --- |
| `key.converter` | CLASS | `org.apache.kafka.connect.storage.StringConverter` | Key deserializer. StringConverter for plain string keys, JsonConverter for JSON keys. Using the wrong converter is the #1 cause of task failures. |
| `value.converter` | CLASS | `org.apache.kafka.connect.json.JsonConverter` | Value deserializer. JsonConverter for JSON, AvroConverter for Avro, StringConverter for plain text. |
| `value.converter.schemas.enable` | BOOLEAN | `false` | Whether JSON messages include a schema wrapper. Set false for plain JSON objects. |
| `offset.flush.interval.ms` | LONG | `10000` | Offset commit interval in milliseconds. |

## Required Parameters

These must be set in `connector_config` for the connector to start.

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `topics` | LIST | (none) | Comma-separated list of Kafka topics to consume. |
| `s3.bucket.name` | STRING | (none) | S3 bucket name for data output. |
| `s3.region` | STRING | `us-west-2` | AWS region of the S3 bucket. Validator: valid AWS region. |
| `storage.class` | CLASS | (none) | Storage implementation. Fixed: `io.confluent.connect.s3.storage.S3Storage`. |
| `format.class` | CLASS | (none) | Output format class. See format options table below. |

### Format Options

| format.class | Description | File Extension |
| --- | --- | --- |
| `io.confluent.connect.s3.format.json.JsonFormat` | JSON format (recommended for most use cases) | `.json` |
| `io.confluent.connect.s3.format.avro.AvroFormat` | Apache Avro binary format | `.avro` |
| `io.confluent.connect.s3.format.parquet.ParquetFormat` | Apache Parquet columnar format | `.snappy.parquet` |
| `io.confluent.connect.s3.format.bytearray.ByteArrayFormat` | Raw bytes passthrough | configurable (default `.bin`) |

## Recommended Parameters

| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `flush.size` | INT | (no default) | Records per partition before flushing to S3. Recommended: 1000. |
| `rotate.interval.ms` | LONG | (no default) | Time-based flush interval in ms. Recommended: 600000 (10 min). |
