# Confluent vs AutoMQ Connector Terraform API 设计对比

## 1. 概述

本文对比 Confluent Cloud 和 AutoMQ Connect 多租模型下的 Terraform 资源设计。当前 AutoMQ Provider 已采用两层资源模型：

- `automq_connect_cluster`：管理 BYOC 环境中的 Kafka Connect Worker 集群、插件、容量、计算资源和 Worker 配置。
- `automq_connector`：管理挂载到某个 Connect Cluster 上的业务 Connector 实例。

这意味着旧单资源模型中的插件、K8s、容量、IAM、Worker 配置和指标导出字段不再属于 `automq_connector`。

## 2. 核心架构差异

### Confluent Cloud：全托管 SaaS 模式

Confluent 的 `confluent_connector` 资源代表一个 connector 实例。Connect Worker 集群由 Confluent 托管，用户不需要配置 worker 数量、K8s 部署或资源规格。

### AutoMQ：BYOC 多租模式

AutoMQ 将 Worker 集群和 Connector 实例拆成两个资源。用户通过 `automq_connect_cluster` 管理 BYOC 基础设施，通过 `automq_connector` 在已有集群上创建业务 connector。

这个拆分让多个 connector 复用同一个 Connect Cluster，插件安装和容量调整也能独立于 connector 生命周期演进。

## 3. 参数对比

### Confluent `confluent_connector`

| 参数 | 类型 | 说明 |
|------|------|------|
| `environment.id` | Required | 环境 ID |
| `kafka_cluster.id` | Required | Kafka 集群 ID |
| `config_nonsensitive` | Required, Map | 非敏感 connector 配置 |
| `config_sensitive` | Required, Map | 敏感 connector 配置 |
| `status` | Optional | 状态控制 |
| `offsets` | Optional | Offset 管理 |

关键设计点：

- connector 类型通常通过配置项声明。
- capacity、worker、K8s 等基础设施参数不暴露给用户。
- 认证信息也放在 connector 配置体系内。

### AutoMQ `automq_connect_cluster`

| 参数 | 类型 | 说明 |
|------|------|------|
| `environment_id` | Required | 环境 ID |
| `name` | Required | Connect Cluster 名称 |
| `plugins` | Required | Worker 集群安装的插件名称和版本 |
| `kafka_cluster.kafka_instance_id` | Required | Worker 集群使用的 AutoMQ Kafka 实例 |
| `capacity` | Required | Worker 容量模式和规格 |
| `compute` | Required | K8s/ASG 等计算资源配置 |
| `worker_config` | Optional | Worker 级配置 |
| `metric_exporter` | Optional | Worker 指标导出配置 |
| `version` | Optional | Worker 版本 |

### AutoMQ `automq_connector`

| 参数 | 类型 | 说明 |
|------|------|------|
| `environment_id` | Required | 环境 ID |
| `connect_cluster_id` | Required | 运行该 connector 的 Connect Cluster |
| `name` | Required | Connector 名称 |
| `connector_class` | Required | Java connector class |
| `task_count` | Required | task 数量 |
| `kafka_cluster.security_protocol` | Optional | connector 插件内部 producer/consumer 的 Kafka 鉴权 |
| `connector_config` | Optional | 非敏感业务配置 |
| `connector_config_sensitive` | Optional | 敏感业务配置 |
| `initial_offsets` | Optional | 创建时初始化 offset |

`automq_connector` 的只读字段包括 `id`、`state`、`connector_type`、`plugin_id`、`created_at` 和 `updated_at`。其中 `plugin_id` 由后端根据 `connect_cluster_id` 和 `connector_class` 解析，不是用户输入。

## 4. `plugin_id` 设计结论

旧模型把 `plugin_id` 放在 `automq_connector` 顶层，因为当时一个 connector 资源同时负责部署 Worker 集群和创建 connector 实例。

多租模型下这个设计已经不成立：

- 插件安装属于 Worker 集群能力，应由 `automq_connect_cluster.plugins` 声明。
- Connector 实例只需要声明 `connector_class`，后端在目标 cluster 已安装插件中解析对应插件。
- `plugin_id` 保留为 computed 字段，方便观察和排查，但不能作为 `automq_connector` 入参。

因此，最终 Terraform API 中 `automq_connector` 不应暴露 `plugin_id`、`plugin_type`、K8s、IAM、capacity、worker_config、metric_exporter、version 或 scheduling_spec 等 Worker 集群字段。

## 5. 总结

| 对比项 | Confluent | AutoMQ 多租模型 |
|--------|-----------|----------------|
| Worker 集群 | 平台托管，不暴露 | `automq_connect_cluster` 显式管理 |
| Connector 实例 | `confluent_connector` | `automq_connector` |
| 插件选择 | connector 配置或平台内置能力 | `automq_connect_cluster.plugins` |
| `plugin_id` | 不是 connector 顶层参数 | `automq_connector.plugin_id` 只读 |
| 基础设施参数 | 不暴露 | 只在 `automq_connect_cluster` 暴露 |
| 业务配置 | config map | `connector_config` / `connector_config_sensitive` |

核心结论：AutoMQ 多租 Terraform API 的边界应和产品模型一致。`automq_connect_cluster` 管 Worker 集群与插件基础设施，`automq_connector` 只管业务 connector 实例。
