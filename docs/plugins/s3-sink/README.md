# S3 Sink Connector — 插件沉淀进度

## 状态：PoC 验证通过，源码分析完成，配置模板和文档已生成

## PoC 结果

- 环境：CMP `http://3.0.58.151:8080`，env-zjtouprrltmx6591
- 插件：conn-plugin-be3329c4（automq-kafka-connect-s3-11.1.0.zip）
- Kafka 实例：kf-setsleeo7t8dujsf
- Connector：状态 RUNNING，10 条消息端到端流动验证通过

### PoC 中发现的配置要点

1. Worker 的 `key.converter` 默认是 `JsonConverter`，如果 message key 不是 JSON 会导致 task FAILED
2. 正确配置：`key.converter=org.apache.kafka.connect.storage.StringConverter`
3. `value.converter.schemas.enable=false`（如果 value 是无 schema 的 JSON）
4. 这些配置应该放在 workerConfig 中

### PoC 使用的配置

```json
{
  "workerConfig": {
    "key.converter": "org.apache.kafka.connect.storage.StringConverter",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false"
  },
  "connectorConfig": {
    "topics": "s3-sink-poc-topic",
    "s3.region": "ap-southeast-1",
    "s3.bucket.name": "demo-connect-test1-automq",
    "flush.size": "3",
    "storage.class": "io.confluent.connect.s3.storage.S3Storage",
    "format.class": "io.confluent.connect.s3.format.json.JsonFormat",
    "partitioner.class": "io.confluent.connect.storage.partitioner.DefaultPartitioner",
    "rotate.interval.ms": "60000",
    "topics.dir": "s3-sink-poc"
  }
}
```

## 源码分析（进行中）

源码位置：`/tmp/kafka-connect-s3/kafka-connect-s3/`
主类：`io.confluent.connect.s3.S3SinkConnector`
配置类：`io.confluent.connect.s3.S3SinkConnectorConfig`（1505 行，41 个 S3 特有参数）

### S3 特有配置参数

| 参数 | 说明 |
|------|------|
| `s3.bucket.name` | S3 bucket 名称（必填） |
| `s3.region` | S3 region |
| `s3.part.size` | 分片上传大小 |
| `s3.compression.type` | 压缩类型（none/gzip） |
| `s3.ssea.name` | 服务端加密算法 |
| `s3.sse.kms.key.id` | KMS key ID |
| `s3.acl.canned` | S3 ACL |
| `s3.wan.mode` | WAN 模式 |
| `s3.proxy.url` | 代理 URL |
| `s3.credentials.provider.class` | 凭证提供者 |
| `aws.access.key.id` | AWS Access Key（敏感） |
| `aws.secret.access.key` | AWS Secret Key（敏感） |
| `format.class` | 输出格式（JsonFormat/AvroFormat/ParquetFormat/ByteArrayFormat） |
| `format.bytearray.extension` | ByteArray 格式扩展名 |
| `format.bytearray.separator` | ByteArray 格式分隔符 |
| `partitioner.class` | 分区策略 |
| `flush.size` | 触发 flush 的记录数 |
| `rotate.interval.ms` | 时间触发 flush |
| `store.kafka.keys` | 是否存储 key |
| `store.kafka.headers` | 是否存储 headers |
| `behavior.on.null.values` | null 值处理策略 |
| `json.decimal.format` | JSON decimal 格式 |

## 待完成

- [x] 完整的 ConfigDef 参数表（含类型、默认值、是否必填、是否敏感）
- [x] 配置模板生成（config-template.json）
- [x] Quick Start 文档
- [x] Configuration Reference 文档
- [x] Performance Tuning 文档（含基准测试数据）
- [x] Migration Guide 文档（Confluent + MSK）
- [x] Troubleshooting 文档
- [x] PR 提交
- [ ] 更大规模性能测试（不同 Tier 对比）
- [ ] CMP 指标面板截图

## 性能基准（初步）

- 环境：TIER1（0.5 CPU / 1 GiB），1 Worker，1 Task
- Produce 吞吐：80 msgs/sec（通过 CMP API，含 HTTP 开销）
- Connector 状态：RUNNING，无错误
- flush.size=3，100 条消息全部成功处理
