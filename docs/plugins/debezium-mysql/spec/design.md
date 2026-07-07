# Debezium MySQL CDC Source Connector 沉淀设计

## 技术方案

### 1. MySQL 环境

使用 Docker 本地启动 MySQL 8.0，开启 binlog：
```bash
docker run -d --name mysql-debezium \
  -e MYSQL_ROOT_PASSWORD=debezium \
  -e MYSQL_USER=debezium \
  -e MYSQL_PASSWORD=debezium \
  -e MYSQL_DATABASE=testdb \
  -p 3306:3306 \
  mysql:8.0 \
  --server-id=1 \
  --log-bin=mysql-bin \
  --binlog-format=ROW \
  --binlog-row-image=FULL \
  --gtid-mode=ON \
  --enforce-gtid-consistency=ON
```

### 2. PoC 验证流程

1. 启动 MySQL Docker 容器
2. 创建测试表和初始数据
3. 创建 Service Account（如果不存在）
4. 创建 Kafka Instance（复用已有或新建）
5. 创建 Kafka Topic（用于接收 CDC 数据）
6. 创建 Debezium MySQL Source Connector
7. 在 MySQL 中执行 INSERT/UPDATE/DELETE
8. 验证 Kafka Topic 中收到对应的 CDC 事件

### 3. Connector 配置（基于 Debezium 3.1 文档）

核心配置：
```json
{
  "connector.class": "io.debezium.connector.mysql.MySqlConnector",
  "database.hostname": "host.docker.internal",
  "database.port": "3306",
  "database.user": "debezium",
  "database.password": "debezium",
  "database.server.id": "184054",
  "topic.prefix": "mysql-cdc",
  "database.include.list": "testdb",
  "schema.history.internal.kafka.bootstrap.servers": "${kafka.bootstrap.servers}",
  "schema.history.internal.kafka.topic": "schema-changes.testdb"
}
```

### 4. 性能测试方案

Source Connector 的性能测试与 Sink 不同：
- 测试指标：source records polled/s、source records written/s、端到端延迟
- 测试方法：在 MySQL 中批量插入数据，观察 Connector 的消费速率
- 配置变量：snapshot.mode、max.batch.size、poll.interval.ms

### 5. 文档结构

```
docs/plugins/debezium-mysql/
├── README.md                    # 概述
├── quick-start.md               # 快速开始
├── configuration-reference.md   # 配置参考
├── performance-tuning.md        # 性能调优
├── migration-guide.md           # 迁移指南
└── troubleshooting.md           # 故障排查
```

## 风险与注意事项

1. MySQL binlog 权限：需要 REPLICATION SLAVE、REPLICATION CLIENT 权限
2. 网络连通性：Docker 容器内的 MySQL 需要能被 K8s 中的 Connector 访问
   - 方案 A：使用公网 IP 的 MySQL（如 RDS）
   - 方案 B：在 K8s 集群内部署 MySQL Pod
3. Schema History Topic：Debezium 需要一个额外的 topic 存储 schema 变更历史
