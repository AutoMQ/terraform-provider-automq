# Debezium MySQL CDC Source Connector - Troubleshooting

## 常见问题及解决方案

### 1. Connector 启动失败

#### 1.1 KafkaSchemaHistory 配置错误

**错误信息**：
```
org.apache.kafka.connect.errors.ConnectException: Error configuring an instance of KafkaSchemaHistory; check the logs for details
```

**原因**：缺少 `schema.history.internal.kafka.bootstrap.servers` 配置

**解决方案**：
添加完整的 schema.history.internal 配置：
```json
{
  "schema.history.internal.kafka.topic": "schema-changes.your_database",
  "schema.history.internal.kafka.bootstrap.servers": "your-kafka:9102",
  "schema.history.internal.consumer.security.protocol": "SASL_PLAINTEXT",
  "schema.history.internal.consumer.sasl.mechanism": "SCRAM-SHA-512",
  "schema.history.internal.consumer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"user\" password=\"pass\";",
  "schema.history.internal.producer.security.protocol": "SASL_PLAINTEXT",
  "schema.history.internal.producer.sasl.mechanism": "SCRAM-SHA-512",
  "schema.history.internal.producer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"user\" password=\"pass\";"
}
```

#### 1.2 SASL 认证失败

**错误信息**：
```
org.apache.kafka.common.errors.IllegalSaslStateException: Unexpected handshake request with client mechanism SCRAM-SHA-512, enabled mechanisms are []
```

**原因**：使用了错误的 Kafka 端口（9092 是 PLAINTEXT，9102 是 SASL_PLAINTEXT）

**解决方案**：
- 确认 Kafka 端口配置正确
- SASL_PLAINTEXT 通常使用 9102 端口
- 检查 `schema.history.internal.kafka.bootstrap.servers` 端口

#### 1.3 MySQL 连接失败

**错误信息**：
```
io.debezium.DebeziumException: Unable to connect: Communications link failure
```

**原因**：
- MySQL 主机不可达
- 防火墙阻止连接
- MySQL 未启动

**解决方案**：
1. 验证 MySQL 主机可达：`telnet mysql-host 3306`
2. 检查防火墙规则
3. 确认 MySQL 服务运行中

### 2. 权限问题

#### 2.1 MySQL 复制权限不足

**错误信息**：
```
Access denied; you need (at least one of) the REPLICATION SLAVE privilege(s) for this operation
```

**解决方案**：
```sql
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'debezium'@'%';
GRANT SELECT, RELOAD, SHOW DATABASES, LOCK TABLES ON *.* TO 'debezium'@'%';
FLUSH PRIVILEGES;
```

#### 2.2 Kafka ACL 权限不足

**错误信息**：
```
org.apache.kafka.common.errors.GroupAuthorizationException: Not authorized to access group: connect-xxx
```

**解决方案**：
为 Kafka 用户添加以下 ACL：
- TOPIC: ALL（LITERAL: *）
- GROUP: ALL（PREFIXED: connect）
- CLUSTER: ALL
- TRANSACTIONAL_ID: ALL（PREFIXED: connect）

### 3. Binlog 问题

#### 3.1 Binlog 未启用

**错误信息**：
```
The MySQL server is not configured to use a ROW binlog_format
```

**解决方案**：
在 MySQL 配置文件中启用 binlog：
```ini
[mysqld]
server-id=1
log-bin=mysql-bin
binlog-format=ROW
binlog-row-image=FULL
```

重启 MySQL 后验证：
```sql
SHOW VARIABLES LIKE 'log_bin';
SHOW VARIABLES LIKE 'binlog_format';
```

#### 3.2 Binlog 位置丢失

**错误信息**：
```
The connector is trying to read binlog starting at ... but this is no longer available on the server
```

**原因**：MySQL binlog 已被清理，Connector 记录的位置不存在

**解决方案**：
1. 重置 Connector offset（删除 Kafka Connect 内部 topic 中的 offset）
2. 使用 `snapshot.mode=initial` 重新执行快照
3. 或使用 `snapshot.mode=recovery` 恢复

### 4. 性能问题

#### 4.1 快照执行缓慢

**症状**：初始快照耗时过长

**解决方案**：
1. 增加 `snapshot.fetch.size`（默认 2000）
2. 增加 `snapshot.max.threads`（并行快照）
3. 使用 `table.include.list` 限制表范围
4. 考虑使用 `snapshot.mode=no_data` 跳过数据快照

#### 4.2 CDC 延迟高

**症状**：binlog 消费延迟增加

**解决方案**：
1. 减小 `poll.interval.ms`（默认 500ms）
2. 增加 `max.batch.size`（默认 2048）
3. 增加 `max.queue.size`（默认 8192）
4. 升级 Worker Tier（增加 CPU/内存）

### 5. 数据问题

#### 5.1 时间戳精度丢失

**症状**：毫秒/微秒精度丢失

**解决方案**：
设置 `time.precision.mode=adaptive_time_microseconds`

#### 5.2 Decimal 精度问题

**症状**：Decimal 类型数据精度不正确

**解决方案**：
- `decimal.handling.mode=precise`：使用 BigDecimal（精确但序列化复杂）
- `decimal.handling.mode=double`：使用 double（可能丢失精度）
- `decimal.handling.mode=string`：使用字符串（推荐）

### 6. 诊断命令

#### 检查 Connector 状态
```bash
# 通过 CMP API
curl -s "http://cmp:8080/api/v1/connectors/{connector-id}" | jq '.state, .taskStates'

# 通过 Connect REST API（在 Pod 内）
kubectl exec -it {connect-pod} -- curl -s http://localhost:8083/connectors/{connector-name}/status
```

#### 检查 Connector 日志
```bash
kubectl logs {connect-pod} --tail=100 | grep -E "ERROR|WARN|Exception"
```

#### 检查 MySQL binlog 状态
```sql
SHOW MASTER STATUS;
SHOW BINARY LOGS;
SHOW VARIABLES LIKE 'binlog%';
```

#### 检查 Kafka Topic
```bash
# 列出 topic
kafka-topics.sh --bootstrap-server kafka:9102 --list | grep mysql-cdc

# 查看 topic 消息数
kafka-run-class.sh kafka.tools.GetOffsetShell \
  --broker-list kafka:9102 \
  --topic mysql-cdc.db.table
```

## 日志级别调整

如需更详细的日志，可以在 Connector 配置中添加：
```json
{
  "log.level": "DEBUG"
}
```

或者通过 Connect REST API 动态调整：
```bash
curl -X PUT http://localhost:8083/admin/loggers/io.debezium \
  -H "Content-Type: application/json" \
  -d '{"level": "DEBUG"}'
```
