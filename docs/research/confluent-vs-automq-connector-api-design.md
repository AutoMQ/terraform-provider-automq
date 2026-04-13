# Confluent vs AutoMQ Connector Terraform API 设计对比

## 1. 概述

本文对比 Confluent Cloud 和 AutoMQ 的 Connector Terraform 资源设计，重点分析 `pluginId` 参数的合理性，以及两者在架构理念上的根本差异。

**信息来源**：
- [Confluent Terraform Provider - confluent_connector](https://www.pulumi.com/registry/packages/confluentcloud/api-docs/connector/) (Content was rephrased for compliance with licensing restrictions)
- [Confluent Terraform Provider - confluent_custom_connector_plugin](https://www.pulumi.com/registry/packages/confluentcloud/api-docs/customconnectorplugin/) (Content was rephrased for compliance with licensing restrictions)

## 2. 核心架构差异

### Confluent Cloud：全托管 SaaS 模式

Confluent 的 `confluent_connector` 资源代表的是**一个 connector 实例**（单个 connector task），而不是一个 Connect 集群。Confluent 完全托管了 Connect 集群的生命周期，用户不需要关心 worker 数量、K8S 部署、资源规格等基础设施细节。

### AutoMQ：BYOC 自管模式

AutoMQ 的 `automq_connector` 资源代表的是**一个 Connect 集群**（包含多个 worker pod），部署在用户自己的 K8S 集群上。用户需要管理 worker 数量、资源规格、K8S namespace、ServiceAccount、IAM Role 等基础设施配置。

## 3. 参数对比

### Confluent `confluent_connector` 的参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `environment.id` | Required | 环境 ID |
| `kafka_cluster.id` | Required | Kafka 集群 ID |
| `config_nonsensitive` | Required, Map | 所有非敏感配置（扁平 KV map） |
| `config_sensitive` | Required, Map | 所有敏感配置（扁平 KV map） |
| `status` | Optional | 状态控制（RUNNING/PAUSED） |
| `offsets` | Optional | Offset 管理 |

关键设计点：
- **没有 `plugin_id` 参数**。connector 类型通过 `config_nonsensitive` 中的 `connector.class` 键指定
- **没有独立的 capacity/worker 参数**。Confluent 全托管，用户只需指定 `tasks.max`
- **没有 K8S 相关参数**。基础设施完全由 Confluent 管理
- **没有独立的 security_protocol 嵌套块**。认证通过 `kafka.auth.mode` + `kafka.service.account.id` 或 `kafka.api.key` 在 config map 中配置

### Confluent `confluent_custom_connector_plugin` 的参数

Confluent 将插件管理拆分为独立资源：

| 参数 | 类型 | 说明 |
|------|------|------|
| `display_name` | Required | 插件显示名 |
| `connector_class` | Required | Java 全限定类名 |
| `connector_type` | Required | SOURCE 或 SINK |
| `filename` | Required | 插件 jar/zip 文件路径 |
| `cloud` | Optional | 云提供商（AWS/AZURE/GCP） |
| `sensitive_config_properties` | Optional | 敏感配置属性名列表 |

使用自定义插件时，在 `confluent_connector` 的 config 中通过以下方式引用：
```hcl
config_nonsensitive = {
  "confluent.connector.type"   = "CUSTOM"
  "confluent.custom.plugin.id" = confluent_custom_connector_plugin.my_plugin.id
  "connector.class"            = confluent_custom_connector_plugin.my_plugin.connector_class
}
```

### AutoMQ `automq_connector` 的参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `environment_id` | Required | 环境 ID |
| `name` | Required | 集群显示名 |
| `plugin_id` | Required | 插件 ID（系统内置或用户上传） |
| `plugin_type` | Optional | SOURCE 或 SINK |
| `connector_class` | Optional | Java 全限定类名 |
| `kubernetes_cluster_id` | Required | K8S 集群 ID |
| `kubernetes_namespace` | Required | K8S namespace |
| `kubernetes_service_account` | Required | K8S ServiceAccount |
| `iam_role` | Optional | AWS IAM Role ARN (IRSA) |
| `task_count` | Required | 任务并行数 |
| `capacity` | Required | worker 数量 + 资源规格 |
| `kafka_cluster` | Required | Kafka 实例 ID + 安全协议配置 |
| `worker_config` | Optional | Worker 级配置覆盖 |
| `connector_config` | Optional | Connector 级配置 |
| `metric_exporter` | Optional | 指标导出配置 |

## 4. `plugin_id` 设计分析

### Confluent 的做法

Confluent **不使用 plugin_id 作为 connector 的参数**。原因：

1. **内置 connector**：通过 `connector.class` 的简短别名引用（如 `DatagenSource`、`S3_SINK`），不需要 ID
2. **自定义 connector**：通过 `confluent.custom.plugin.id` 在 config map 中引用，这是一个普通的配置项而非顶层参数
3. **关注点分离**：插件是独立资源（`confluent_custom_connector_plugin`），connector 只是使用它

### AutoMQ 的做法

AutoMQ 将 `plugin_id` 作为 connector 的**顶层 Required 参数**。原因：

1. AutoMQ 的 Connect 集群需要知道加载哪个插件来启动 worker
2. 插件决定了 connector 的类型和可用的 connector class
3. 在 BYOC 模式下，插件需要被部署到 K8S pod 中，这是集群级别的配置

### 合理性评估

**AutoMQ 的 `plugin_id` 设计是合理的**，原因如下：

| 维度 | Confluent | AutoMQ | 分析 |
|------|-----------|--------|------|
| 抽象层级 | connector 实例 | Connect 集群 | AutoMQ 管理的是集群，插件是集群级配置 |
| 插件部署 | 平台托管 | 用户 K8S | 需要明确指定哪个插件部署到 pod |
| 插件来源 | 内置 + 自定义上传 | 系统内置 + 用户上传 | 两者类似 |
| 引用方式 | config map 中的 KV | 顶层参数 | AutoMQ 更显式，Confluent 更灵活 |

但有几个可以改进的点：

### 改进建议

#### 建议 1：考虑是否需要独立的 `automq_connector_plugin` 资源

目前 `plugin_id` 引用的是后端预注册的插件。如果未来支持用户通过 Terraform 上传自定义插件，应该像 Confluent 一样拆分为独立资源。

Confluent 实际上有完整的插件版本管理体系，拆分为两个资源：
- `confluent_custom_connector_plugin` — 插件本体（display_name、connector_class、connector_type）
- `confluent_custom_connector_plugin_version` — 插件版本（version、filename、plugin_id）

每个版本可以上传不同的 jar/zip，并关联到同一个 plugin。参考：[CustomConnectorPluginVersion](https://www.pulumi.com/registry/packages/confluentcloud/api-docs/customconnectorpluginversion/) (Content was rephrased for compliance with licensing restrictions)

```hcl
# Confluent 的自定义插件版本管理示例
resource "confluent_custom_connector_plugin" "datagen" {
  display_name    = "Datagen Source"
  connector_class = "io.confluent.kafka.connect.datagen.DatagenConnector"
  connector_type  = "SOURCE"
  filename        = "confluentinc-kafka-connect-datagen-0.6.2.zip"
}

resource "confluent_custom_connector_plugin_version" "v1" {
  plugin_id  = confluent_custom_connector_plugin.datagen.id
  version    = "v1.2.4"
  filename   = "confluentinc-kafka-connect-datagen-0.6.2.zip"
  # ...
}
```

AutoMQ 未来可考虑类似设计：

```hcl
# 未来可能的设计
resource "automq_connector_plugin" "s3_sink" {
  environment_id  = "env-xxx"
  display_name    = "S3 Sink Connector"
  connector_class = "io.confluent.connect.s3.S3SinkConnector"
  connector_type  = "SINK"
  filename        = "./plugins/confluentinc-kafka-connect-s3-10.5.0.zip"
}

resource "automq_connector" "test" {
  plugin_id = automq_connector_plugin.s3_sink.id
  # ...
}
```

#### 建议 2：`connector_class` 是否应该从 plugin 自动推导

当前 `connector_class` 是 Optional 的，如果 plugin 只包含一个 connector class，后端可以自动推导。这个设计是合理的。但如果 plugin 包含多个 class（如 Debezium 包含 MySQL/PostgreSQL/MongoDB source），则 `connector_class` 变为必需。

建议在文档中明确说明：当插件包含多个 connector class 时，`connector_class` 为必填。

#### 建议 3：`config_sensitive` / `config_nonsensitive` 分离模式

Confluent 将所有配置拍平到两个 map 中（sensitive 和 nonsensitive），这种设计的优点：
- 极度灵活，支持任意 connector 的任意配置
- 不需要为每种 connector 定义不同的 schema
- 新增配置项不需要修改 provider 代码

AutoMQ 当前的设计将部分配置提升为顶层参数（`kafka_cluster.security_protocol`），部分放在 `connector_config` map 中。这种混合模式的优点是类型安全和文档友好，缺点是灵活性较低。

**当前设计是合理的**，因为 AutoMQ 管理的是 Connect 集群而非单个 connector 实例，安全协议是集群级配置，提升为结构化参数更合适。

## 5. 总结

| 对比项 | Confluent | AutoMQ | 结论 |
|--------|-----------|--------|------|
| 资源粒度 | connector 实例 | Connect 集群 | 不同产品形态，各自合理 |
| `plugin_id` | 无（在 config 中引用） | 顶层 Required | AutoMQ 合理，因为插件是集群级配置 |
| 插件版本控制 | 内置 connector 无法指定版本；自定义插件有独立的 version 资源 | 通过 `plugin_id` + `version` 控制 | AutoMQ 更统一，内置和自定义插件都可控 |
| 配置模式 | 全部扁平 KV map | 结构化 + KV map 混合 | AutoMQ 更类型安全，Confluent 更灵活 |
| 基础设施参数 | 无 | K8S/capacity/IAM | AutoMQ 必需，BYOC 模式决定 |
| 安全认证 | config map 中的 KV | 结构化嵌套块 | AutoMQ 更友好，有 schema 校验 |
| 插件管理 | 独立资源（plugin + plugin_version） | 引用已有 ID | AutoMQ 未来可考虑拆分为独立资源 |

**核心结论**：AutoMQ 的 `plugin_id` 设计是合理的。两个产品的架构差异（全托管 vs BYOC）决定了 API 设计的不同方向。Confluent 追求极致灵活性（一切皆 config），AutoMQ 追求类型安全和显式声明（结构化参数）。两种方式各有优劣，AutoMQ 当前的选择适合其 BYOC 定位。
