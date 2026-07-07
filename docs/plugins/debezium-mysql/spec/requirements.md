# Debezium MySQL CDC Source Connector 沉淀需求

## 目标

完成 Debezium MySQL CDC Source Connector 的完整沉淀，包括：
1. PoC 验证（端到端数据流动）
2. 源码分析（ConfigDef 参数提取）
3. 性能基准测试
4. 配置模板生成
5. 文档产出

## 插件信息

- 插件名称：debezium-mysql
- 版本：3.1.2
- 类型：SOURCE
- CDN URL：http://download.automq.com/resource/connector/debezium-debezium-connector-mysql-3.1.2.zip
- 官方文档：https://debezium.io/documentation/reference/connectors/mysql.html
- GitHub：https://github.com/debezium/debezium

## 环境信息

- CMP 地址：http://3.0.58.151:8080
- 环境 ID：env-zjtouprrltmx6591
- Admin 账号：admin / gaopp121212*
- IAM Role：arn:aws:iam::780817326912:role/keqing-admin

## 外部依赖

MySQL 数据库（需要开启 binlog）：
- 可以用 Docker 本地启动：`docker run mysql:8.0 --server-id=1 --log-bin=mysql-bin --binlog-format=ROW`
- 或者使用 AWS RDS MySQL（需要用户提供）

## 验收标准

1. PoC 验证：MySQL 表变更 → Debezium Connector → Kafka Topic，数据端到端流动
2. 源码分析：完整的 ConfigDef 参数表 + 特有指标列表
3. 性能测试：不同配置组合下的吞吐/延迟数据 + 可视化图表
4. 文档：Quick Start、Configuration Reference、Performance Tuning、Migration Guide、Troubleshooting
