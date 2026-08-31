# Debezium MySQL CDC Source Connector - Configuration Reference

## 核心配置参数

### 数据库连接

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| database.hostname | string | 是 | - | MySQL 服务器地址 |
| database.port | int | 否 | 3306 | MySQL 端口 |
| database.user | string | 是 | - | MySQL 用户名（需要复制权限） |
| database.password | string | 是 | - | MySQL 密码 |
| database.server.id | int | 是 | - | 唯一的 server ID，用于 binlog 复制 |

### Topic 配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| topic.prefix | string | 是 | - | Kafka topic 前缀，用于生成 topic 名称 |
| database.include.list | string | 否 | - | 要捕获的数据库列表（逗号分隔） |
| database.exclude.list | string | 否 | - | 要排除的数据库列表 |
| table.include.list | string | 否 | - | 要捕获的表列表（格式：db.table） |
| table.exclude.list | string | 否 | - | 要排除的表列表 |

### Schema History 配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| schema.history.internal.kafka.topic | string | 是 | - | 存储 schema 变更历史的 topic |
| schema.history.internal.kafka.bootstrap.servers | string | 是 | - | Kafka bootstrap servers |
| schema.history.internal.consumer.security.protocol | string | 否 | PLAINTEXT | Consumer 安全协议 |
| schema.history.internal.consumer.sasl.mechanism | string | 否 | - | Consumer SASL 机制 |
| schema.history.internal.consumer.sasl.jaas.config | string | 否 | - | Consumer JAAS 配置 |
| schema.history.internal.producer.security.protocol | string | 否 | PLAINTEXT | Producer 安全协议 |
| schema.history.internal.producer.sasl.mechanism | string | 否 | - | Producer SASL 机制 |
| schema.history.internal.producer.sasl.jaas.config | string | 否 | - | Producer JAAS 配置 |

### Snapshot 配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| snapshot.mode | string | 否 | initial | 快照模式 |
| snapshot.locking.mode | string | 否 | minimal | 快照锁定模式 |
| snapshot.fetch.size | int | 否 | 2000 | 快照时每次获取的行数 |
| snapshot.max.threads | int | 否 | 1 | 快照并行线程数 |

#### snapshot.mode 可选值

| 值 | 说明 |
|----|------|
| initial | 首次启动时执行快照，之后只读取 binlog |
| initial_only | 只执行快照，不读取 binlog |
| when_needed | 需要时执行快照 |
| never | 从不执行快照，只读取 binlog |
| no_data | 只快照 schema，不快照数据 |
| recovery | 恢复模式，用于 schema history 丢失时 |

#### snapshot.locking.mode 可选值

| 值 | 说明 |
|----|------|
| minimal | 最小锁定，只在读取 schema 时持有全局读锁 |
| extended | 扩展锁定，整个快照期间持有全局读锁 |
| none | 不使用锁（仅用于 schema_only 模式） |

### 数据转换配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| key.converter | string | 否 | JsonConverter | Key 转换器 |
| value.converter | string | 否 | JsonConverter | Value 转换器 |
| key.converter.schemas.enable | boolean | 否 | true | Key 是否包含 schema |
| value.converter.schemas.enable | boolean | 否 | true | Value 是否包含 schema |
| decimal.handling.mode | string | 否 | precise | Decimal 处理模式 |
| time.precision.mode | string | 否 | adaptive_time_microseconds | 时间精度模式 |
| bigint.unsigned.handling.mode | string | 否 | long | BIGINT UNSIGNED 处理模式 |

### 性能配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| max.batch.size | int | 否 | 2048 | 每批次最大事件数 |
| max.queue.size | int | 否 | 8192 | 内部队列最大大小 |
| poll.interval.ms | long | 否 | 500 | 轮询间隔（毫秒） |
| heartbeat.interval.ms | int | 否 | 0 | 心跳间隔（0 表示禁用） |

### SSL/TLS 配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| database.ssl.mode | string | 否 | preferred | SSL 模式 |
| database.ssl.keystore | string | 否 | - | Keystore 路径 |
| database.ssl.keystore.password | string | 否 | - | Keystore 密码 |
| database.ssl.truststore | string | 否 | - | Truststore 路径 |
| database.ssl.truststore.password | string | 否 | - | Truststore 密码 |

#### database.ssl.mode 可选值

| 值 | 说明 |
|----|------|
| disabled | 不使用 SSL |
| preferred | 优先使用 SSL，不可用时回退到非加密 |
| required | 必须使用 SSL |
| verify_ca | 使用 SSL 并验证 CA |
| verify_identity | 使用 SSL 并验证主机名 |

### GTID 配置

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| gtid.source.includes | string | 否 | - | 包含的 GTID 源 UUID（正则表达式） |
| gtid.source.excludes | string | 否 | - | 排除的 GTID 源 UUID（正则表达式） |

## 完整配置示例

```json
{
  "name": "mysql-cdc-production",
  "connectorClass": "io.debezium.connector.mysql.MySqlConnector",
  "taskCount": 1,
  "connectorConfig": {
    "properties": {
      "database.hostname": "mysql.example.com",
      "database.port": "3306",
      "database.user": "debezium",
      "database.password": "${secrets:mysql-password}",
      "database.server.id": "184054",
      "topic.prefix": "prod-mysql",
      "database.include.list": "orders,inventory",
      "table.include.list": "orders.orders,orders.order_items,inventory.products",

      "schema.history.internal.kafka.topic": "schema-changes.prod-mysql",
      "schema.history.internal.kafka.bootstrap.servers": "kafka:9102",
      "schema.history.internal.consumer.security.protocol": "SASL_PLAINTEXT",
      "schema.history.internal.consumer.sasl.mechanism": "SCRAM-SHA-512",
      "schema.history.internal.consumer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"debezium\" password=\"${secrets:kafka-password}\";",
      "schema.history.internal.producer.security.protocol": "SASL_PLAINTEXT",
      "schema.history.internal.producer.sasl.mechanism": "SCRAM-SHA-512",
      "schema.history.internal.producer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"debezium\" password=\"${secrets:kafka-password}\";",

      "snapshot.mode": "initial",
      "snapshot.locking.mode": "minimal",
      "snapshot.fetch.size": "10000",

      "key.converter": "org.apache.kafka.connect.json.JsonConverter",
      "value.converter": "org.apache.kafka.connect.json.JsonConverter",
      "key.converter.schemas.enable": "false",
      "value.converter.schemas.enable": "false",

      "max.batch.size": "4096",
      "max.queue.size": "16384",
      "poll.interval.ms": "100",
      "heartbeat.interval.ms": "10000",

      "database.ssl.mode": "required"
    }
  }
}
```

## 敏感参数

以下参数包含敏感信息，在 CMP 中会被加密存储：

- `database.password`
- `database.ssl.keystore.password`
- `database.ssl.truststore.password`
- `schema.history.internal.consumer.sasl.jaas.config`
- `schema.history.internal.producer.sasl.jaas.config`
