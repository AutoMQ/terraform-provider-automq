---
inclusion: auto
---

# AutoMQ Connector Terraform API v2 — 完整上下文

当用户提到 connect 的 terraform API、connector API 重新设计、API v2 等相关话题时，自动加载本文件。

## 项目背景

这是 `terraform-provider-automq` 项目，AutoMQ BYOC Kafka 的 Terraform Provider。
后端代码在 `automqbox/` 目录（symlink 到 CMP Spring Boot 项目，分支 `develop/202601`）。
Terraform Provider 分支：`feat/connector-resource`。

## 已完成的工作

### 1. automq_connector 资源（已实现，已合入 PR #106）
- 完整 CRUD + ImportState + 异步等待
- Schema：environment_id, name, plugin_id(Required), connector_class(Required), plugin_type, kubernetes_cluster_id, kubernetes_namespace, kubernetes_service_account, iam_role, task_count, capacity, kafka_cluster(含 security_protocol), worker_config, connector_config, metric_exporter, labels, version, scheduling_spec
- Bug 修复：sasl_mechanism/truststore_certs 需要 Optional+Computed（后端自动填充）
- E2E 测试：自包含依赖链 instance→user→topic→ACLs→connector，约 19 分钟

### 2. automq_connector_plugin 资源（已实现）
- 不可变资源：Create + Read + Delete，无 Update
- Schema：environment_id, name, version, storage_url, types, connector_class, description, documentation_link
- Bug 修复：GET 返回 200+空 body 视为 404；connector_class 从 sinkConnectorClasses/sourceConnectorClasses 提取；provider 是保留字改为 plugin_provider
- E2E 测试：约 9 秒

### 3. connector_class 改为 Required
- 后端 ConnectorCreateParam 中 connectorClass 无 @NotBlank 注解，但业务上必填
- 后端没有根据 plugin 自动选择 connector class 的逻辑

### 4. Confluent vs AutoMQ API 调研文档
- 位置：`docs/research/confluent-vs-automq-connector-api-design.md`
- 核心结论：Confluent 全托管用扁平 KV map，AutoMQ BYOC 用结构化参数，各自合理

## 会议决定（Phase 3 差异化）

### 核心命题
2026 年核心命题："承接 MSK Connect / Confluent Cloud 迁移到 AutoMQ 的客户需求，让 Connect 不成为客户迁移的阻碍。"

### 会议结论
1. **API 重新设计是最高优先级**，需考虑单租多租模式和部署方式（K8s + ASG）
2. **内置插件产品化**：插件实体存在但配置脱离，内置插件屏蔽 plugin_id，参考 Confluent 简化体验
3. **Offset 设置支持**：通过 KIP-875 offset management API，Terraform API 和 UI 都透出
4. **ASG 部署模式**：与 K8s 并行支持
5. **Top 插件配置简化**：梳理常用插件，沉淀默认配置模板
6. **Table Topic 联动**：先不考虑

### 客户迁移分析

**MSK 客户**：
- 感知 Worker，容量模型（MCU × Worker）与 AutoMQ（Tier × Worker）自然映射
- 插件都是 custom plugin，需重新上传
- 区分 worker 配置和插件配置
- 额外需要补充 K8s 部署参数
- Auto-Scaling 模式 AutoMQ 暂不支持

**Confluent 客户**：
- 极简 API（只有 config_nonsensitive + config_sensitive 两个 KV map）
- 不感知 Worker、插件、K8s、IAM
- connector.class 用简称（如 MySqlCdcSource），需映射为 Java 全限定类名
- 迁移工作量比 MSK 大得多

## API v2 设计文档

位置：`docs/design/automq-connector-terraform-api-v2.md`

### 核心变更（相对现状）

| 变更 | 现状 | 终态 | 理由 |
|------|------|------|------|
| plugin_id | Required 入参 | Computed 只读 | 用 connector_class + plugin_version 替代，后端自动匹配 |
| plugin_type | Optional 入参 | 移除 | 后端从插件元数据推导 SOURCE/SINK |
| connector_type | 不存在 | Computed 只读（新增） | 后端推导后返回 |
| plugin_version | 不存在 | Optional+Computed（新增） | BYOC 模式需要版本控制，省略用最新版本 |
| K8s 参数 | 顶层 Required | 移入 compute.kubernetes 块 | 支持多部署模式 |
| compute 块 | 不存在 | Required（新增） | 抽象 K8s/ASG |
| iam_role | 顶层 Optional | 移入 compute.iam_role | IAM 是计算资源属性 |
| scheduling_spec | 顶层 Optional | 移入 compute.kubernetes.scheduling_spec | K8s 特有 |
| connector_config_sensitive | 不存在 | Optional（新增） | 敏感配置分离 |
| initial_offsets | 不存在 | Optional（新增） | Offset 迁移（KIP-875） |
| security_protocol.security_protocol | 同名嵌套 | 改为 security_protocol.protocol | 消除困惑 |
| key_password | Optional | 移除 | 前端未暴露，冗余 |

关键设计决策：用 connector_class + plugin_version 替代 plugin_id。
- Confluent 是全托管 SaaS，不需要版本控制
- AutoMQ 是 BYOC，插件部署在用户侧，需要 plugin_version 控制升级节奏
- connector_class 是必填，plugin_version 可选（默认最新）
- connector_type 不需要用户声明，后端从插件元数据推导

### 后端 API 依赖

| 改动 | 优先级 |
|------|--------|
| connector_class + plugin_version 匹配逻辑 | P0 |
| plugin_id 改为可选/只读 | P0 |
| connector_type 自动推导 | P0 |
| security_protocol.protocol 字段名调整 | P0 |
| connector_config_sensitive 支持 | P1 |
| initial_offsets 支持（KIP-875） | P1 |
| ASG 部署模式 | P2 |

## 环境信息

- CMP endpoint: `http://13.229.117.145:8080`
- Environment ID: `env-urumxrrhrljo61a9`
- K8S cluster: `eks-41fdf-automqlab`
- Service Account: access_key=`RYL56iOF1Jf5iTmJ`
- 测试插件: `conn-plugin-e866cccb`（S3 Sink，storageUrl: `http://download.automq.com/resource/connector/automq-kafka-connect-s3-11.1.0.zip`）
- acc-config: `test-connector/acc-config.json`
- ~/.terraformrc 跑测试前需要备份
- golangci-lint: `$(go env GOPATH)/bin/golangci-lint`

## 后端代码位置（automqbox，分支 develop/202601）

CMP 是 Spring Boot 项目，通过 symlink `automqbox/` 引用。

| 模块 | 路径 |
|------|------|
| Connector Controller | `automqbox/cmp/cmp-app/src/main/java/com/automq/cmp/controller/ConnectorController.java` |
| Connector Plugin Controller | `automqbox/cmp/cmp-app/src/main/java/com/automq/cmp/controller/ConnectorPluginController.java` |
| Connector Service | `automqbox/cmp/cmp-service/src/main/java/com/automq/cmp/service/impl/ConnectorServiceImpl.java` |
| Plugin Service | `automqbox/cmp/cmp-service/src/main/java/com/automq/cmp/service/impl/ConnectPluginServiceImpl.java` |
| Plugin Manager | `automqbox/cmp/cmp-service/src/main/java/com/automq/cmp/manager/impl/ConnectPluginManagerImpl.java` |
| Connector Create Param | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/param/connect/ConnectorCreateParam.java` |
| Plugin Create Param | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/param/connect/ConnectPluginCreateParam.java` |
| Connector VO | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/vo/connect/ConnectorVO.java` (在 ConnectorAssembler 中组装) |
| Plugin VO | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/vo/connect/ConnectPluginVO.java` |
| Plugin Entity | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/entity/connect/ConnectPlugin.java` |
| Connector Entity | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/entity/connect/KubernetesConnector.java` |
| Plugin State/Type/Provider Enums | `automqbox/cmp/cmp-common/src/main/java/com/automq/cmp/model/entity/connect/PluginState.java`, `PluginType.java`, `PluginProvider.java` |
| 前端 Connect 页面 | `automqbox/cmp/cmp-frontend-next/src/pages/connect/` |
| 前端认证字段组件 | `automqbox/cmp/cmp-frontend-next/src/components/fields/authentication-fields.tsx` |

## 设计文档位置

| 文档 | 路径 |
|------|------|
| API v2 详细设计 | `docs/design/automq-connector-terraform-api-v2.md` |
| API v2 Schema 速览（含多租方案对比） | `docs/design/automq-connector-api-v2-schema.md` |
| 项目规划（12 周） | `docs/design/connect-project-plan.md` |
| Confluent vs AutoMQ 调研 | `docs/research/confluent-vs-automq-connector-api-design.md` |

## 当前状态

API v2 设计文档已完成（含多租方案对比），等待用户评审。
项目规划文档已完成：`docs/design/connect-project-plan.md`。

### 多租户重构进行中

- Issue: https://github.com/AutoMQ/automqbox/issues/5491
- AIP 文档: https://automq66.feishu.cn/wiki/FxhHwMitQi58RKk8JwHcrrL9nwe
- Spec: `.kiro/specs/connect-multi-tenant/`（requirements + design + tasks）
- 后端分支: `feat/connect-multi-tenant`（从 `develop/202603` 切出）
- Terraform API Schema: `docs/design/automq-connector-api-v2-schema.md`（评审后终版）

#### 鉴权信息拆分（设计变更）

security_protocol（username/password 等鉴权信息）从 `automq_connect_cluster.kafka_cluster` 移到 `automq_connector.kafka_cluster`：
- ConnectCluster 的 Worker 进程连接 Kafka broker 使用 CMP 内部鉴权，不需要用户提供
- Connector 插件内部创建的 producer/consumer client 需要用户提供鉴权信息
- `automq_connect_cluster.kafka_cluster` 只保留 `kafka_instance_id`
- `automq_connector` 新增 `kafka_cluster.security_protocol` 块

#### CMP 数据库迁移机制

CMP 使用 `SimpleDDLExecutor`（非 Flyway），启动时自动扫描 `classpath*:sql/**/*.sql` 按文件名排序执行。
- 每条 SQL 独立执行，失败只 warn 不中断
- 所有迁移 SQL 必须幂等（`CREATE TABLE IF NOT EXISTS`、`INSERT OR IGNORE`、`WHERE ... IS NULL`）
- BYOC 客户升级时 CMP 重启即自动执行
- SQL 文件命名规则：`v{version}_{description}.sql`（如 `v8.3.0_connect_multi_tenant.sql`）
- 现有最高版本：v8.2.0（已被 batch_operation、connect_version_default_worker_config、pay_as_you_go 占用）
- 现有 SQL 文件位置：`cmp-app/src/main/resources/sql/`

### S3 Sink 插件沉淀进行中

CMP 环境信息（本次沉淀专用）：
- Endpoint: `http://3.0.58.151:8080`
- Console login: admin / gaopp121212*
- Login API: `POST /api/v1/auth`（返回 Set-Cookie accessToken/refreshToken）
- Service Account: access_key=`4xN9iWBLeknfGMk6`, secret_key=`boIyIcOrWPvkW0IUHcarOS7hOD11tRD8`
- Environment ID: `env-zjtouprrltmx6591`
- Region: ap-southeast-1
- S3 bucket: `demo-connect-test1-automq`（ap-southeast-1）
- IAM Role: `arn:aws:iam::780817326912:role/keqing-admin`
- 已有 7 个内置插件（system-plugins.json），S3 Sink 在 CDN 上但未注册

SOP 进度：
- 步骤 1-4：完成（选型、适配、PoC、源码分析）
- 步骤 5（性能测试）：
  - Kafka produce 吞吐已测（native producer via kubectl）：200B=65K rec/sec, 500B=54K rec/sec, 1KB=40K rec/sec
  - Connector 消费吞吐测试失败：Spot 实例不稳定导致 connector 创建超时或运行后立即 FAILED
  - 需要稳定环境（非 Spot 实例）才能完成 connector 消费吞吐的多配置对比测试
- 步骤 6-7：文档和配置模板已完成，benchmark 数据待补充 connector 消费部分
- 步骤 8：PR 已提交（produce 吞吐数据 + 图表）

待解决：
1. 需要稳定的非 Spot 实例环境来完成 connector 消费吞吐测试
2. kafka-producer-perf-test 发送的是 raw bytes，connector 需要用 StringConverter + ByteArrayFormat 才能处理
3. 或者用 kafka-console-producer 发送 JSON 格式消息，connector 用 JsonConverter + JsonFormat

SOP 执行中发现的关键问题和解决方案：

1. **新环境需要创建 Service Account**：
   - 登录 `POST /api/v1/auth` → 创建 SA `POST /api/v1/service-accounts`
   - SA 需要 username + roleBindings（role=EnvironmentAdmin）

2. **K8s 验证失败（Connector.KubernetesValidationFailed）**：
   - 需要用户在 EKS 中添加 CMP 的访问条目
   - 需要确认安全组已开放
   - 源码分析：`validateKubernetesPrerequisites` 用 kafkaInstance 的 isNetworkInternal()，IAAS 实例可能为 null/false，导致 refreshKubeConfig 跳过

3. **Produce Message API**：
   - 路径：`POST /api/v1/instances/{instanceId}/topics/{topicId}/message-channels`
   - topicId 必须用 Base64 编码的内部 ID（如 `1gL5o_7QQTum13zEFEMVxA`），不是 topic name
   - Body: `{"messages":[{"key":"xxx","content":"xxx"}]}`
   - 查询 topic ID：`GET /api/v1/instances/{instanceId}/topics?page=1&size=10`

4. **S3 Sink Connector 配置关键点**：
   - Worker 的 key.converter 默认是 JsonConverter，如果 message key 不是 JSON 会导致 task FAILED
   - 正确配置：`key.converter=org.apache.kafka.connect.storage.StringConverter`
   - `value.converter.schemas.enable=false`（如果 value 是无 schema 的 JSON）
   - 这些配置应该放在 workerConfig 中，不是 connectorConfig

## 插件沉淀 SOP

详见独立 steering 文件：`.kiro/steering/plugin-onboarding-sop.md`

## AIP 工作流

AutoMQ 内部所有功能开发前都要走 AIP 流程：
- 状态：Proposed → Accepted → WIP → Pre-Release → Released
- 模板章节：背景、问题定义、调研论证、解决方案、原型设计、接口设计、依赖选型、方案详情、兼容性、被拒绝方案、落地计划、验收
- 一个 AIP 从开始到关闭设计为 1-4 周
