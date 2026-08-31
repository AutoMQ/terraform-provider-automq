# Debezium MySQL CDC Source Connector PoC 配置

## 验证时间
2026-04-02

## 环境信息
- CMP 地址：http://3.0.58.151:8080
- 环境 ID：env-zjtouprrltmx6591
- Kafka Instance：kf-fcbb5380s8hv7iwt
- K8s Cluster：eks-1jy59-automqlab
- Connector ID：automq-connect-318d4f4f

## MySQL 环境
- 部署方式：K8s Pod（mysql:8.0）
- Service：mysql-debezium.default.svc.cluster.local:3306
- 用户：debezium / debezium
- 数据库：testdb
- 测试表：customers (id, name, email, created_at)
- Binlog 配置：ROW 格式，GTID 模式

## Connector 配置

```json
{
  "name": "debezium-mysql-poc-v3",
  "kubernetesClusterId": "eks-1jy59-automqlab",
  "pluginId": "conn-plugin-f9bd2ea6",
  "connectorClass": "io.debezium.connector.mysql.MySqlConnector",
  "type": "SOURCE",
  "taskCount": 1,
  "kubernetesNamespace": "default",
  "kubernetesServiceAccount": "kafka-connect-sa",
  "capacity": {
    "workerCount": 1,
    "workerResourceSpec": "TIER1"
  },
  "kafkaCluster": {
    "kafkaInstanceId": "kf-fcbb5380s8hv7iwt",
    "securityProtocolConfig": {
      "securityProtocol": "SASL_PLAINTEXT",
      "username": "debezium-user",
      "password": "DebeziumPoC123!",
      "saslMechanism": "SCRAM-SHA-512"
    }
  },
  "connectorConfig": {
    "properties": {
      "database.hostname": "mysql-debezium.default.svc.cluster.local",
      "database.port": "3306",
      "database.user": "debezium",
      "database.password": "debezium",
      "database.server.id": "184054",
      "topic.prefix": "mysql-cdc",
      "database.include.list": "testdb",
      "schema.history.internal.kafka.topic": "schema-changes.testdb",
      "schema.history.internal.kafka.bootstrap.servers": "0-kf-fcbb5380s8hv7iwt.nj1ba7u426owwxmm.automq.private:9102",
      "schema.history.internal.consumer.security.protocol": "SASL_PLAINTEXT",
      "schema.history.internal.consumer.sasl.mechanism": "SCRAM-SHA-512",
      "schema.history.internal.consumer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"debezium-user\" password=\"DebeziumPoC123!\";",
      "schema.history.internal.producer.security.protocol": "SASL_PLAINTEXT",
      "schema.history.internal.producer.sasl.mechanism": "SCRAM-SHA-512",
      "schema.history.internal.producer.sasl.jaas.config": "org.apache.kafka.common.security.scram.ScramLoginModule required username=\"debezium-user\" password=\"DebeziumPoC123!\";",
      "key.converter": "org.apache.kafka.connect.json.JsonConverter",
      "value.converter": "org.apache.kafka.connect.json.JsonConverter",
      "key.converter.schemas.enable": "false",
      "value.converter.schemas.enable": "false"
    }
  }
}
```

## 关键配置说明

### 必填配置
| 参数 | 说明 | 示例值 |
|------|------|--------|
| database.hostname | MySQL 主机地址 | mysql-debezium.default.svc.cluster.local |
| database.port | MySQL 端口 | 3306 |
| database.user | MySQL 用户名 | debezium |
| database.password | MySQL 密码 | debezium |
| database.server.id | MySQL server ID（必须唯一） | 184054 |
| topic.prefix | Kafka topic 前缀 | mysql-cdc |
| database.include.list | 要捕获的数据库列表 | testdb |
| schema.history.internal.kafka.topic | Schema 历史 topic | schema-changes.testdb |
| schema.history.internal.kafka.bootstrap.servers | Schema 历史 Kafka 地址 | 0-xxx:9102 |

### Schema History 认证配置
Debezium 需要单独配置 schema.history.internal 的 Kafka 认证：
- `schema.history.internal.consumer.security.protocol`
- `schema.history.internal.consumer.sasl.mechanism`
- `schema.history.internal.consumer.sasl.jaas.config`
- `schema.history.internal.producer.security.protocol`
- `schema.history.internal.producer.sasl.mechanism`
- `schema.history.internal.producer.sasl.jaas.config`

## 验证结果

### CDC 数据样例（Snapshot）
```json
{
  "before": null,
  "after": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2026-04-02T11:09:27Z"
  },
  "source": {
    "version": "3.1.2.Final",
    "connector": "mysql",
    "name": "mysql-cdc",
    "ts_ms": 1775132462000,
    "snapshot": "first",
    "db": "testdb",
    "table": "customers",
    "server_id": 0,
    "gtid": null,
    "file": "mysql-bin.000003",
    "pos": 1606
  },
  "op": "r",
  "ts_ms": 1775132463071
}
```

### CDC 数据样例（Incremental INSERT）
```json
{
  "before": null,
  "after": {
    "id": 4,
    "name": "David",
    "email": "david@example.com",
    "created_at": "2026-04-02T12:21:31Z"
  },
  "source": {
    "version": "3.1.2.Final",
    "connector": "mysql",
    "name": "mysql-cdc",
    "ts_ms": 1775132491005,
    "snapshot": "false",
    "db": "testdb",
    "table": "customers",
    "server_id": 1,
    "gtid": "3cf91c85-2e84-11f1-90ab-26cf6462cbc2:14",
    "file": "mysql-bin.000003",
    "pos": 1832
  },
  "op": "c",
  "ts_ms": 1775132491016
}
```

## 遇到的问题及解决方案

### 1. Task FAILED: KafkaSchemaHistory 配置错误
**错误**：`Error configuring an instance of KafkaSchemaHistory`
**因**：缺少 `schema.history.internal.kafka.bootstrap.servers` 配置
**解决**：添加完整的 schema.history.internal 配置

### 2. Task FAILED: SASL 认证失败
**错误**：`Unexpected handshake request with client mechanism SCRAM-SHA-512, enabled mechanisms are []`
**原因**：schema.history.internal 连接使用了错误的端口（9092 是 PLAINTEXT，9102 是 SASL_PLAINTEXT）
**解决**：使用 9102 端口并配置完整的 SASL 认证

### 3. Consumer 权限不足
**错误**：`Not authorized to access group: console-consumer-xxx`
**原因**：Kafka User 缺少 console-consumer group 的 ACL
**解决**：添加 GROUP ACL（PREFIXED: console-consumer）
