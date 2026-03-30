# S3 Sink Connector — Migration Guide

## 从 Confluent Cloud 迁移

### 参数映射

| Confluent Cloud 参数 | AutoMQ 参数 | 说明 |
|---------------------|-------------|------|
| `connector.class = S3_SINK` | `connector_class = io.confluent.connect.s3.S3SinkConnector` | Confluent 用简称，AutoMQ 用全限定类名 |
| `output.data.format = JSON` | `format.class = io.confluent.connect.s3.format.json.JsonFormat` | 格式映射（见下方完整映射表） |
| `output.data.format = AVRO` | `format.class = io.confluent.connect.s3.format.avro.AvroFormat` | |
| `output.data.format = PARQUET` | `format.class = io.confluent.connect.s3.format.parquet.ParquetFormat` | |
| `input.data.format` | 不需要 | Confluent 专有参数 |
| `kafka.auth.mode` | 不需要 | 认证通过 `kafka_cluster.security_protocol` 配置 |
| `kafka.api.key` | 不需要 | 替换为 SASL username |
| `kafka.api.secret` | 不需要 | 替换为 SASL password |
| `tasks.max` | `task_count` | 直接复制 |
| `topics` | `connector_config["topics"]` | 直接复制 |
| `s3.bucket.name` | `connector_config["s3.bucket.name"]` | 直接复制 |
| `s3.region` | `connector_config["s3.region"]` | 直接复制 |
| `flush.size` | `connector_config["flush.size"]` | 直接复制 |
| `rotate.interval.ms` | `connector_config["rotate.interval.ms"]` | 直接复制 |
| `s3.compression.type` | `connector_config["s3.compression.type"]` | 直接复制 |

### Confluent 专有参数（不需要迁移）

以下参数是 Confluent Cloud 专有的，在 AutoMQ 中不需要：

- `kafka.auth.mode` — Confluent 用 API Key 认证
- `kafka.api.key` / `kafka.api.secret` — 替换为 SASL 用户名密码
- `kafka.service.account.id` — Confluent 服务账号
- `input.data.format` — Confluent 专有
- `output.data.format` — 映射为 `format.class`

### AutoMQ 额外需要的参数

Confluent Cloud 是全托管服务，以下参数在 Confluent 中不需要但在 AutoMQ 中必须配置：

| 参数 | 说明 |
|------|------|
| `storage.class` | 固定值 `io.confluent.connect.s3.storage.S3Storage` |
| `key.converter` | Worker 配置，推荐 `StringConverter` |
| `value.converter` | Worker 配置，根据消息格式选择 |
| `capacity` | Worker 数量和规格（Confluent 不暴露） |
| `compute` | K8s/ASG 部署参数（Confluent 不需要） |
| `kafka_cluster.security_protocol` | Kafka 认证配置 |

### 迁移步骤

1. 从 Confluent Cloud 导出 connector 配置（`config_nonsensitive` + `config_sensitive`）
2. 将 `connector.class` 从 `S3_SINK` 映射为 `io.confluent.connect.s3.S3SinkConnector`
3. 将 `output.data.format` 映射为 `format.class`
4. 移除 Confluent 专有参数（`kafka.auth.mode`, `kafka.api.key` 等）
5. 添加 `storage.class = io.confluent.connect.s3.storage.S3Storage`
6. 配置 Worker converter（`key.converter`, `value.converter`）
7. 决定容量规格（参考 Performance Tuning Guide）
8. 配置 K8s/ASG 部署参数
9. 创建 SASL 用户替代 API Key
10. 通过 Terraform 创建 connector

## 从 MSK Connect 迁移

### 参数映射

| MSK Connect 参数 | AutoMQ 参数 | 说明 |
|-----------------|-------------|------|
| `connector.class` | `connector_class` | 直接复制（MSK 用全限定类名） |
| `tasks.max` | `task_count` | 直接复制 |
| `topics` | `connector_config["topics"]` | 直接复制 |
| 其他 connector 配置 | `connector_config` | 直接复制 |
| `capacity.mcuCount` | `capacity.worker_resource_spec` | MCU → Tier 映射 |
| `capacity.workerCount` | `capacity.worker_count` | 直接复制 |
| `plugins[].customPlugin.arn` | 需要重新上传插件 | 通过 `automq_connector_plugin` 上传 |
| `serviceExecutionRoleArn` | `compute.iam_role` | 直接复制 |
| `workerConfiguration` | `worker_config` | 从 MSK worker_configuration 资源提取 KV |

### MCU → Tier 映射

| MSK MCU | AutoMQ Tier | CPU | 内存 |
|---------|-------------|-----|------|
| 1 MCU | TIER1 | 0.5 | 1 GiB |
| 2 MCU | TIER2 | 1 | 2 GiB |
| 4 MCU | TIER3 | 2 | 4 GiB |
| 8 MCU | TIER4 | 4 | 8 GiB |

### 迁移步骤

1. 从 MSK Connect 导出 connector 配置（`connectorConfiguration`）
2. 上传自定义插件到 AutoMQ（`automq_connector_plugin`）
3. 将 MSK 的 `connectorConfiguration` 拆分为 `connector_config` 和 `worker_config`
4. 将 MCU 映射为 Tier
5. 配置 K8s/ASG 部署参数
6. 配置 Kafka 认证（MSK 可能用 IAM，AutoMQ 用 SASL）
7. 通过 Terraform 创建 connector
