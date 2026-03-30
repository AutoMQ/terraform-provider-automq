# S3 Sink Connector — Configuration Reference

## Worker 配置

这些参数在 `worker_config` 中设置，影响整个 Connect Worker 的行为。

| 参数 | 类型 | 推荐值 | 说明 |
|------|------|--------|------|
| `key.converter` | CLASS | `org.apache.kafka.connect.storage.StringConverter` | Key 序列化器。如果 message key 是普通字符串必须用 StringConverter，否则 JsonConverter 会报错 |
| `value.converter` | CLASS | `org.apache.kafka.connect.json.JsonConverter` | Value 序列化器。JSON 消息用 JsonConverter，Avro 用 AvroConverter |
| `value.converter.schemas.enable` | BOOLEAN | `false` | JSON 消息是否包含 schema wrapper。无 schema 的 JSON 设为 false |
| `offset.flush.interval.ms` | LONG | `10000` | Offset 提交间隔（毫秒） |

## Connector 配置 — 必填参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `topics` | LIST | 要消费的 Kafka topic 列表，逗号分隔 |
| `s3.bucket.name` | STRING | S3 bucket 名称 |
| `s3.region` | STRING | S3 bucket 所在 AWS region |
| `storage.class` | CLASS | 固定值 `io.confluent.connect.s3.storage.S3Storage` |
| `format.class` | CLASS | 输出格式类（见下方格式选项） |

### 格式选项

| format.class | 说明 | 文件扩展名 |
|-------------|------|-----------|
| `io.confluent.connect.s3.format.json.JsonFormat` | JSON 格式（推荐） | `.json` |
| `io.confluent.connect.s3.format.avro.AvroFormat` | Avro 格式 | `.avro` |
| `io.confluent.connect.s3.format.parquet.ParquetFormat` | Parquet 格式 | `.snappy.parquet` |
| `io.confluent.connect.s3.format.bytearray.ByteArrayFormat` | 原始字节 | 自定义 |

## Connector 配置 — 推荐参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `flush.size` | INT | 无默认 | 每个 partition 累积多少条记录后 flush 到 S3。推荐 1000 |
| `rotate.interval.ms` | LONG | 无默认 | 时间触发 flush（毫秒）。推荐 600000（10 分钟） |
| `partitioner.class` | CLASS | `DefaultPartitioner` | 分区策略。DefaultPartitioner 按 Kafka partition，TimeBasedPartitioner 按时间 |
| `topics.dir` | STRING | `topics` | S3 中的顶层目录名 |
| `s3.compression.type` | STRING | `none` | 压缩类型：`none` 或 `gzip` |

## Connector 配置 — 可选参数

### S3 连接

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `s3.wan.mode` | BOOLEAN | `false` | 使用 S3 Transfer Acceleration |
| `s3.path.style.access.enabled` | BOOLEAN | `false` | 使用 path-style 访问（非 virtual-hosted） |
| `s3.proxy.url` | STRING | — | S3 代理 URL |
| `s3.proxy.user` | STRING | — | 代理用户名 |
| `s3.proxy.password` | PASSWORD | — | 代理密码（敏感） |
| `s3.part.size` | INT | `26214400` (25MB) | 分片上传大小（最小 5MB） |
| `s3.retry.backoff.ms` | LONG | `200` | S3 请求重试退避时间 |
| `s3.part.retries` | INT | `3` | 分片上传重试次数 |

### 加密

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `s3.ssea.name` | STRING | — | 服务端加密算法：`AES256` 或 `aws:kms` |
| `s3.sse.kms.key.id` | STRING | — | KMS key ID（`aws:kms` 时必填） |
| `s3.sse.customer.key` | PASSWORD | — | SSE-C 客户端密钥（敏感） |

### 认证

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `aws.access.key.id` | STRING | — | AWS Access Key（敏感）。推荐使用 IAM Role 而非 Access Key |
| `aws.secret.access.key` | PASSWORD | — | AWS Secret Key（敏感） |
| `s3.credentials.provider.class` | CLASS | `DefaultAWSCredentialsProviderChain` | 凭证提供者。使用 IAM Role 时保持默认 |

### 数据处理

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `store.kafka.keys` | BOOLEAN | `false` | 是否在输出文件中包含 message key |
| `store.kafka.headers` | BOOLEAN | `false` | 是否在输出文件中包含 message headers |
| `behavior.on.null.values` | STRING | `fail` | null 值处理：`fail`（报错）、`ignore`（跳过）、`write`（写入 tombstone） |
| `s3.object.tagging` | BOOLEAN | `false` | 是否给 S3 对象添加 offset 标签 |
| `json.decimal.format` | STRING | `BASE64` | JSON 中 decimal 类型的格式：`BASE64` 或 `NUMERIC` |

### 分区策略

| partitioner.class | 说明 | 目录结构示例 |
|------------------|------|------------|
| `DefaultPartitioner` | 按 Kafka partition | `topics/my-topic/0/` |
| `TimeBasedPartitioner` | 按时间 | `topics/my-topic/year=2026/month=03/day=30/hour=12/` |
| `FieldPartitioner` | 按消息字段值 | `topics/my-topic/region=us-east/` |
| `HourlyPartitioner` | 按小时 | `topics/my-topic/2026/03/30/12/` |
| `DailyPartitioner` | 按天 | `topics/my-topic/2026/03/30/` |

使用 `TimeBasedPartitioner` 时需要额外配置：

```
partitioner.class = io.confluent.connect.storage.partitioner.TimeBasedPartitioner
path.format = 'year'=YYYY/'month'=MM/'day'=dd/'hour'=HH
partition.duration.ms = 3600000
locale = en-US
timezone = UTC
```

## 敏感参数列表

以下参数应放在 `connector_config_sensitive` 中（API v2）或标记为 sensitive：

- `aws.access.key.id`
- `aws.secret.access.key`
- `s3.proxy.password`
- `s3.sse.customer.key`
