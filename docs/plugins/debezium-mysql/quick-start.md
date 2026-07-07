# Debezium MySQL CDC Source Connector - Quick Start

本指南帮助你快速配置 Debezium MySQL CDC Source Connector，将 MySQL 数据库的变更实时同步到 Kafka。

## 前置条件

### MySQL 配置要求

1. **启用 Binlog**（ROW 格式）
```sql
-- 检查 binlog 状态
SHOW VARIABLES LIKE 'log_bin';
SHOW VARIABLES LIKE 'binlog_format';
SHOW MASTER STATUS;
```

MySQL 配置文件（my.cnf）：
```ini
[mysqld]
server-id=1
log-bin=mysql-bin
binlog-format=ROW
binlog-row-image=FULL
```

2. **创建 Debezium 用户**（需要复制权限）
```sql
CREATE USER 'debezium'@'%' IDENTIFIED BY 'your_password';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'debezium'@'%';
GRANT SELECT, RELOAD, SHOW DATABASES, LOCK TABLES ON *.* TO 'debezium'@'%';
FLUSH PRIVILEGES;
```

### Kafka 配置要求

1. 创建 Kafka User 并授予以下 ACL：
   - TOPIC: ALL（用于 CDC 数据和 Schema History）
   - GROUP: ALL（PREFIXED: connect）
   - CLUSTER: ALL
   - TRANSACTIONAL_ID: ALL（PREFIXED: connect）

2. 创建 Schema History Topic（可选，Debezium 会自动创建）

## 最小配置

```json
{
  "name": "mysql-cdc-connector",
  "connectorClass": "io.debezium.connector.mysql.MySqlConnector",
  "taskCount": 1,
  "connectorConfig": {
    "properties": {
      "database.hostname": "your-mysql-host",
      "database.port": "3306",
      "database.user": "debezium",
      "database.password": "your_password",
      "database.server.id": "184054",
      "topic.prefix": "mysql-cdc",
      "database.include.list": "your_database",
      "schema.history.internal.kafka.topic": "schema-changes.your_database",
      "schema.history.internal.kafka.bootstrap.servers": "your-kafka-bootstrap:9102",
      "schema.history.internal.consumer.security.protocol": "SASL_PLAINTEXT",
      "schema.history.internal.consumer.sasl.mechanism": "SCRAM-SHA-512",
      "schema.history.internal.consumer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"your_user\" password=\"your_password\";",
      "schema.history.internal.producer.security.protocol": "SASL_PLAINTEXT",
      "schema.history.internal.producer.sasl.mechanism": "SCRAM-SHA-512",
      "schema.history.internal.producer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"your_user\" password=\"your_password\";"
    }
  }
}
```

## Topic 命名规则

Debezium 会为每个表创建一个 Kafka Topic，命名格式：
```
{topic.prefix}.{database}.{table}
```

例如：`mysql-cdc.testdb.customers`

## 验证 CDC 数据

使用 Kafka Console Consumer 验证数据：
```bash
kafka-console-consumer.sh \
  --bootstrap-server your-kafka:9102 \
  --topic mysql-cdc.your_database.your_table \
  --from-beginning \
  --max-messages 5 \
  --consumer-property security.protocol=SASL_PLAINTEXT \
  --consumer-property sasl.mechanism=SCRAM-SHA-512 \
  --consumer-property "sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username=\"your_user\" password=\"your_password\";"
```

## CDC 事件格式

### INSERT 事件（op: "c"）
```json
{
  "before": null,
  "after": {"id": 1, "name": "Alice", "email": "alice@example.com"},
  "source": {"version": "3.1.2.Final", "connector": "mysql", ...},
  "op": "c",
  "ts_ms": 1775132491016
}
```

### UPDATE 事件（op: "u"）
```json
{
  "before": {"id": 1, "name": "Alice", "email": "alice@example.com"},
  "after": {"id": 1, "name": "Alice Smith", "email": "alice@example.com"},
  "source": {...},
  "op": "u",
  "ts_ms": 1775132491020
}
```

### DELETE 事件（op: "d"）
```json
{
  "before": {"id": 1, "name": "Alice Smith", "email": "alice@example.com"},
  "after": null,
  "source": {...},
  "op": "d",
  "ts_ms": 1775132491025
}
```

### Snapshot 事件（op: "r"）
初始快照时，每条记录都是 "r"（read）操作。

## 常见问题

### 1. Connector 启动失败：KafkaSchemaHistory 配置错误
确保配置了完整的 `schema.history.internal.*` 参数，包括 bootstrap.servers 和认证信息。

### 2. SASL 认证失败
检查 Kafka 端口是否正确（9102 是 SASL_PLAINTEXT，9092 是 PLAINTEXT）。

### 3. 权限不足
确保 MySQL 用户有 REPLICATION SLAVE 和 REPLICATION CLIENT 权限。

## 下一步

- [配置参考](configuration-reference.md) - 完整配置参数说明
- [性能调优](performance-tuning.md) - 优化 CDC 性能
- [故障排查](troubleshooting.md) - 常见问题解决方案
