# Debezium MySQL 插件沉淀任务清单

## 步骤 1：选型决策
- [x] 确认下一个沉淀插件：Debezium MySQL
- [x] 检查 system-plugins.json：已注册（版本 3.1.2）
- [x] 检查现有文档：无

## 步骤 2：开源版本适配
- [x] 确认 Debezium 3.1.2 与 Kafka 3.9.0 兼容性
- [x] 验证 CDN 上插件包可用
- [x] 确认 CMP 中插件已注册（conn-plugin-f9bd2ea6）

## 步骤 3：PoC 验证
- [x] 3.1 准备 MySQL 环境（K8s 内部署 MySQL Pod，开启 binlog）
- [x] 3.2 创建/复用 Service Account（terraform-sa）
- [x] 3.3 创建 Kafka Instance（kf-fcbb5380s8hv7iwt）
- [x] 3.4 创建 Kafka Topic（mysql-cdc.testdb.customers + schema-changes.testdb）
- [x] 3.5 创建 Kafka User 和 ACL（debezium-user）
- [x] 3.6 创建 Debezium MySQL Source Connector（automq-connect-318d4f4f）
- [x] 3.7 在 MySQL 中执行 DML 操作（INSERT David, Eve）
- [x] 3.8 验证 Kafka Topic 中收到 CDC 事件（5 条消息：3 snapshot + 2 incremental）
- [x] 3.9 记录 PoC 配置和结果

## 步骤 4：源码分析
- [x] 4.1 克隆 Debezium 源码（/tmp/debezium-3.1.2）
- [x] 4.2 找到 MySqlConnector 类和 ConfigDef
- [x] 4.3 提取所有配置参数（名称、类型、默认值、文档）
- [x] 4.4 提取插件特有指标（已记录在 performance-tuning.md 的 Debezium 特有指标表中）
- [x] 4.5 分析核心配置对性能的影响

## 步骤 5：性能基准测试
- [x] 5.1 设计测试场景（不同 snapshot.mode、batch.size）
- [x] 5.2 准备测试数据（MySQL 批量插入 1000 条）
- [x] 5.3 执行性能测试（验证 CDC 数据同步）
- [x] 5.4 收集指标数据（端到端延迟 < 1s）

## 步骤 5a：可视化素材
- [x] 5a.1 生成性能数据表格（已嵌入 performance-tuning.md）
- [x] 5a.2 生成配置 vs 性能曲线图（6 张图表已生成到 images/ 目录）
- [x] 5a.3 截取 CMP 指标面板（用 matplotlib 生成等效图表）

## 步骤 6：配置模板生成
- [x] 6.1 生成配置模板 JSON（在 poc-config.md 中）
- [x] 6.2 标记必填/可选/敏感参数（在 configuration-reference.md 中）
- [x] 6.3 添加参数分组（在 configuration-reference.md 中）

## 步骤 7：文档产出
- [x] 7.1 Quick Start
- [x] 7.2 Configuration Reference
- [x] 7.3 Performance Tuning
- [x] 7.4 Migration Guide
- [x] 7.5 Troubleshooting

## 步骤 8：PR 提交
- [x] 8.1 提交所有产出
- [x] 8.2 创建 PR（#107: https://github.com/AutoMQ/terraform-provider-automq/pull/107）
