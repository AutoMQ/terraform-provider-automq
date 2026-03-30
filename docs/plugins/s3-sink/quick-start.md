# S3 Sink Connector — Quick Start

## 概述

S3 Sink Connector 将 Kafka topic 中的数据持续写入 Amazon S3 bucket，支持 JSON、Avro、Parquet 等格式。

## 前置条件

1. 一个运行中的 AutoMQ Kafka 实例
2. 一个 S3 bucket（与 Kafka 实例同 region）
3. IAM Role 具有 S3 写入权限（`s3:PutObject`, `s3:GetBucketLocation`, `s3:ListBucket`）
4. Kafka 用户和 ACL 配置完成

## Terraform 配置示例

```hcl
# 1. 上传 S3 Sink 插件（如果使用内置插件可跳过）
resource "automq_connector_plugin" "s3_sink" {
  environment_id  = "env-xxxxx"
  name            = "s3-sink"
  version         = "11.1.0"
  storage_url     = "http://download.automq.com/resource/connector/automq-kafka-connect-s3-11.1.0.zip"
  types           = ["SINK"]
  connector_class = "io.confluent.connect.s3.S3SinkConnector"
}

# 2. 创建 S3 Sink Connector
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

  # Worker 配置 — converter 设置
  worker_config = {
    "key.converter"                     = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter"                   = "org.apache.kafka.connect.json.JsonConverter"
    "value.converter.schemas.enable"    = "false"
  }

  # Connector 配置 — S3 Sink 参数
  connector_config = {
    "topics"           = "order-events,user-events"
    "s3.bucket.name"   = "my-data-lake"
    "s3.region"        = "us-east-1"
    "flush.size"       = "1000"
    "rotate.interval.ms" = "600000"
    "storage.class"    = "io.confluent.connect.s3.storage.S3Storage"
    "format.class"     = "io.confluent.connect.s3.format.json.JsonFormat"
    "partitioner.class" = "io.confluent.connect.storage.partitioner.DefaultPartitioner"
    "topics.dir"       = "kafka-data"
  }

  timeouts = {
    create = "30m"
    delete = "20m"
  }
}
```

## 关键配置说明

### Worker 配置（重要）

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| `key.converter` | `StringConverter` | 如果 message key 是普通字符串，必须用 StringConverter。默认的 JsonConverter 会在 key 不是 JSON 时导致 task 失败 |
| `value.converter` | `JsonConverter` | JSON 格式消息用 JsonConverter，Avro 用 AvroConverter |
| `value.converter.schemas.enable` | `false` | 如果 JSON 消息不包含 schema wrapper，设为 false |

### Connector 配置

| 参数 | 必填 | 说明 |
|------|------|------|
| `topics` | 是 | 要消费的 topic 列表 |
| `s3.bucket.name` | 是 | S3 bucket 名称 |
| `s3.region` | 是 | S3 bucket 所在 region |
| `storage.class` | 是 | 固定值 `io.confluent.connect.s3.storage.S3Storage` |
| `format.class` | 是 | 输出格式：JsonFormat / AvroFormat / ParquetFormat |
| `flush.size` | 推荐 | 每个 partition 累积多少条记录后 flush 到 S3 |
| `rotate.interval.ms` | 推荐 | 时间触发 flush（毫秒），确保数据及时写入 |

### S3 目录结构

数据写入路径：`s3://{bucket}/{topics.dir}/{topic}/{partition}/`

例如：`s3://my-data-lake/kafka-data/order-events/0/order-events+0+0000000000.json`

## 常见问题

### Task 启动后立即 FAILED

错误日志：`DataException: Converting byte[] to Kafka Connect data failed due to serialization error`

原因：message key 或 value 的格式与 converter 配置不匹配。

解决：检查 `key.converter` 和 `value.converter` 配置是否与实际消息格式一致。

### 数据没有写入 S3

检查：
1. IAM Role 是否有 S3 写入权限
2. `flush.size` 是否设置过大（需要累积足够多的记录才会 flush）
3. `rotate.interval.ms` 是否设置（时间触发 flush）
4. Topic 中是否有数据

## 从 Confluent Cloud 迁移

如果你之前在 Confluent Cloud 上使用 S3 Sink Connector，需要注意：

1. Confluent 的 `output.data.format` 参数需要映射为 `format.class`：
   - `JSON` → `io.confluent.connect.s3.format.json.JsonFormat`
   - `AVRO` → `io.confluent.connect.s3.format.avro.AvroFormat`
   - `PARQUET` → `io.confluent.connect.s3.format.parquet.ParquetFormat`

2. Confluent 专有参数（`kafka.auth.mode`, `kafka.api.key` 等）不需要，认证通过 `kafka_cluster.security_protocol` 配置

3. Confluent 不需要指定 `storage.class`，AutoMQ 需要显式指定
