# Debezium MySQL CDC Source Connector - Migration Guide

本指南帮助你从 Confluent Cloud、Amazon MSK Connect 或自建 Kafka Connect 迁移到 AutoMQ CMP。

## 从 Confluent Cloud 迁移

### 配置映射

| Confluent Cloud 参数 | AutoMQ CMP 参数 | 说明 |
|---------------------|-----------------|------|
| database.hostname | database.hostname | 相同 |
| database.port | database.port | 相同 |
| database.user | database.user | 相同 |
| database.password | database.password | 相同 |
| database.server.id | database.server.id | 相同 |
| database.server.name | topic.prefix | Debezium 3.x 改名 |
| database.whitelist | database.include.list | 已弃用，使用新名称 |
| database.blacklist | database.exclude.list | 已弃用，使用新名称 |
| table.whitelist | table.include.list | 已弃用，使用新名称 |
| table.blacklist | table.exclude.list | 已弃用，使用新名称 |

### Confluent 专有参数

以下参数是 Confluent 专有的，在开源版 Debezium 中不存在：

| Confluent 参数 | 说明 | 替代方案 |
|---------------|------|---------|
| output.data.format | 输出格式 | 使用 key.converter/value.converter |
| output.key.format | Key 格式 | 使用 key.converter |
| output.value.format | Value 格式 | 使用 value.converter |

### 迁移步骤

1. **导出现有配置**
   - 从 Confluent Cloud Console 导出 Connector 配置
   - 记录当前的 offset 位置（如果需要断点续传）

2. **转换配置**
   - 将 `database.server.name` 改为 `topic.prefix`
   - 将 `*.whitelist` 改为 `*.include.list`
   - 将 `*.blacklist` 改为 `*.exclude.list`
   - 添加 `schema.history.internal.*` 配置

3. **创建 AutoMQ Connector**
   - 在 CMP 中创建新的 Connector
   - 使用转换后的配置

4. **验证数据**
   - 检查 CDC 数据是否正常流动
   - 验证数据格式是否一致

## 从 Amazon MSK Connect 迁移

### 配置差异

MSK Connect 使用 AWS 托管的 Kafka Connect，配置方式略有不同：

| MSK Connect | AutoMQ CMP | 说明 |
|-------------|-----------|------|
| 通过 AWS Console 配置 | 通过 CMP API/Console 配置 | 配置方式不同 |
| 使用 IAM 认证 | 使用 SASL 认证 | 认证方式不同 |
| 自动管理 Worker | 手动选择 Worker Tier | 资源管理不同 |

### 迁移步骤

1. **导出 MSK Connect 配置**
   ```bash
   aws kafkaconnect describe-connector --connector-arn <arn> \
     --query 'connectorConfiguration'
   ```

2. **转换认证配置**
   - MSK 使用 IAM 认证，需要改为 SASL 认证
   - 添加 `schema.history.internal.*` 的 SASL 配置

3. **处理 Offset**
   - MSK Connect 的 offset 存储在 MSK 内部 topic
   - 如需断点续传，需要手动迁移 offset

## 从自建 Kafka Connect 迁移

### 迁移步骤

1. **停止现有 Connector**
   ```bash
   curl -X PUT http://connect:8083/connectors/mysql-cdc/pause
   ```

2. **记录 Offset**
   ```bash
   # 查看 offset topic
   kafka-console-consumer.sh \
     --bootstrap-server kafka:9092 \
     --topic connect-offsets \
     --from-beginning \
     --property print.key=true \
     | grep mysql-cdc
   ```

3. **导出配置**
   ```bash
   curl http://connect:8083/connectors/mysql-cdc/config > config.json
   ```

4. **在 CMP 中创建 Connector**
   - 使用导出的配置
   - 添加 CMP 特有的配置（kafkaCluster、capacity 等）

5. **验证并切换**
   - 验证新 Connector 正常运行
   - 停止旧 Connector
   - 删除旧 Connector

## Offset 迁移（断点续传）

如果需要从上次停止的位置继续消费，需要迁移 offset。

### 方法 1：使用 snapshot.mode=never

如果 binlog 位置仍然有效：
```json
{
  "snapshot.mode": "never"
}
```

Connector 会从 binlog 当前位置开始读取。

### 方法 2：手动设置 offset

1. 获取当前 binlog 位置：
   ```sql
   SHOW MASTER STATUS;
   ```

2. 在 Connector 配置中指定起始位置（需要 Debezium 支持）

### 方法 3：重新执行快照

如果 binlog 位置已失效：
```json
{
  "snapshot.mode": "initial"
}
```

这会重新执行完整快照。

## 版本兼容性

### Debezium 版本对照

| AutoMQ CMP 版本 | Debezium 版本 | Kafka Connect 版本 |
|----------------|--------------|-------------------|
| 当前 | 3.1.2 | 3.9.0 |

### 配置兼容性

Debezium 3.x 相比 2.x 有以下重要变更：

1. `database.server.name` → `topic.prefix`
2. `database.history.*` → `schema.history.internal.*`
3. 移除了部分已弃用的参数

## 迁移检查清单

- [ ] 导出现有 Connector 配置
- [ ] 转换配置参数名称
- [ ] 配置 schema.history.internal 认证
- [ ] 创建 Kafka User 和 ACL
- [ ] 创建 AutoMQ Connector
- [ ] 验证 Connector 状态为 RUNNING
- [ ] 验证 CDC 数据正常流动
- [ ] 验证数据格式一致
- [ ] 停止并删除旧 Connector
- [ ] 监控新 Connector 运行状态
