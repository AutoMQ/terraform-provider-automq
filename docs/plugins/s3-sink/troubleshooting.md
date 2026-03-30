# S3 Sink Connector — Troubleshooting

## Task 启动后立即 FAILED

### 症状

Connector 创建后 task 状态变为 FAILED，日志中出现：

```
ERROR WorkerSinkTask Task threw an uncaught and unrecoverable exception
Caused by: org.apache.kafka.connect.errors.DataException: Converting byte[] to Kafka Connect data failed due to serialization error
```

### 原因

Message key 或 value 的格式与 Worker 的 converter 配置不匹配。

最常见的情况：Worker 默认使用 `JsonConverter`，但 message key 是普通字符串（不是 JSON），导致反序列化失败。

### 解决

在 `worker_config` 中配置正确的 converter：

```hcl
worker_config = {
  "key.converter"                  = "org.apache.kafka.connect.storage.StringConverter"
  "value.converter"                = "org.apache.kafka.connect.json.JsonConverter"
  "value.converter.schemas.enable" = "false"
}
```

| 消息格式 | key.converter | value.converter |
|---------|--------------|-----------------|
| Key=字符串, Value=JSON（无 schema） | StringConverter | JsonConverter + schemas.enable=false |
| Key=JSON, Value=JSON（无 schema） | JsonConverter + schemas.enable=false | JsonConverter + schemas.enable=false |
| Key=字符串, Value=JSON（有 schema） | StringConverter | JsonConverter + schemas.enable=true |
| Key=字符串, Value=Avro | StringConverter | AvroConverter |
| Key=字符串, Value=纯文本 | StringConverter | StringConverter |

## 数据没有写入 S3

### 症状

Connector 状态 RUNNING，但 S3 bucket 中没有新文件。

### 排查步骤

1. **检查 topic 中是否有数据**：通过 CMP 的消息查看功能确认 topic 中有消息

2. **检查 flush.size 配置**：如果 `flush.size=1000` 但 topic 中只有 100 条消息，需要等待 `rotate.interval.ms` 触发时间 flush

3. **检查 rotate.interval.ms**：如果没有配置 `rotate.interval.ms`，数据只会在达到 `flush.size` 时写入

4. **检查 IAM 权限**：IAM Role 需要以下 S3 权限：
   - `s3:PutObject`
   - `s3:GetBucketLocation`
   - `s3:ListBucket`
   - `s3:AbortMultipartUpload`（分片上传）
   - `s3:ListMultipartUploadParts`（分片上传）

5. **检查 S3 bucket 和 region**：确认 `s3.bucket.name` 和 `s3.region` 配置正确

## Connector 创建失败：KubernetesValidationFailed

### 症状

```
Error: Create Connector Error
API Error: Connector.KubernetesValidationFailed
Message: Failed to validate Kubernetes prerequisites
```

### 排查步骤

1. **检查 K8s 集群可访问性**：
   ```
   GET /api/v1/providers/k8s-clusters/{clusterId}
   ```
   确认 `accessible: true`

2. **检查 EKS 访问条目**：CMP 需要在 EKS 中有访问权限。联系管理员在 EKS 的 aws-auth ConfigMap 或 Access Entries 中添加 CMP 的 IAM Role

3. **检查安全组**：确认 EKS 节点的安全组允许 CMP 的入站连接

4. **检查 namespace 和 service account**：确认指定的 K8s namespace 和 service account 存在

## S3 写入延迟高

### 症状

数据从 produce 到出现在 S3 的延迟超过预期。

### 优化

1. **减小 flush.size**：更小的 flush.size 意味着更频繁的 S3 写入
2. **配置 rotate.interval.ms**：确保有时间触发 flush
3. **增加 task_count**：更多 task 并行消费
4. **升级 Worker 规格**：CPU 不足会导致处理延迟
5. **启用 S3 Transfer Acceleration**：跨 region 写入时设置 `s3.wan.mode=true`

## Connector 状态 FAILED 但无明显错误

### 排查

1. 通过 CMP 日志 API 查看 connector 日志：
   ```
   GET /api/v1/connectors/{connectorId}/logs?tailLines=100
   ```

2. 检查 Worker pod 的资源使用（CPU/内存是否达到限制）

3. 检查 Kafka 连接是否正常（ACL 是否完整）

### 必需的 ACL

S3 Sink Connector 需要以下 Kafka ACL：

| 资源类型 | 资源名 | 模式 | 操作 |
|---------|--------|------|------|
| TOPIC | * 或具体 topic 名 | LITERAL | ALL |
| GROUP | connect（前缀） | PREFIXED | ALL |
| CLUSTER | kafka-cluster | LITERAL | ALL |
| TRANSACTIONAL_ID | connect（前缀） | PREFIXED | ALL |

缺少任何一个都可能导致 connector 无法正常工作。
