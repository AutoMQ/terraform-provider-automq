# Requirements Document: CMP Connect 多租户重构

## Introduction

CMP Connect 模块从单租户模型重构为多租户模型。当前 CMP 的 Connect 模块采用单资源模型：一个 KubernetesConnector Entity 同时承载 Worker 集群和 Connector 实例，不支持一个集群运行多个 Connector。本次重构将资源模型拆分为两层：`ConnectCluster`（Worker 集群）+ `Connector`（Connector 实例），支持 1:N 多租户共享。

核心命题：承接 MSK Connect / Confluent Cloud 迁移到 AutoMQ 的客户需求，让 Connect 不成为客户迁移的阻碍。

工作范围涵盖三个层面：
1. 后端 CMP（automqbox Spring Boot 项目）：新建 ConnectCluster API，重构 Connector API，数据库迁移
2. 前端 CMP：页面结构调整（集群管理 + Connector 管理分离）
3. Terraform Provider（terraform-provider-automq）：新建 `automq_connect_cluster` 资源，重构 `automq_connector` 资源

第一步产出 AIP（AutoMQ Improvement Proposal）文档，经评审后进入开发。

## Glossary

- **CMP**: AutoMQ Cloud Management Platform，AutoMQ 的云管理平台
- **ConnectCluster**: 重构后的 Worker 集群资源，管理 K8s Deployment、容量、Kafka 连接、插件集
- **Connector**: 重构后的 Connector 实例资源，挂载到 ConnectCluster 上运行，管理 connector class、task count、connector 配置
- **Worker**: Kafka Connect Worker 进程，运行在 K8s Pod 中，负责执行 Connector Task
- **Plugin**: Kafka Connect 插件包，包含 Connector 实现类（如 S3SinkConnector、MySqlConnector）
- **Connect_REST_API**: Kafka Connect 框架提供的 REST API，用于管理 Connector 实例的生命周期
- **AIP**: AutoMQ Improvement Proposal，AutoMQ 内部大项目的设计评审文档
- **Terraform_Provider**: HashiCorp Terraform 的 AutoMQ Provider，通过 HCL 声明式管理 AutoMQ 资源
- **Strimzi**: 开源 Kafka on Kubernetes 项目，本次插件管理策略参考 Strimzi 的 cluster 级别插件声明模式
- **HPA**: Kubernetes Horizontal Pod Autoscaler，用于 autoscaling 容量模式下的 Worker Pod 自动伸缩
- **Internal_Topic**: Kafka Connect 内部使用的 Topic（config、offset、status），用于存储 Connector 配置、消费位点和状态
- **Rebalance**: Kafka Connect Worker 集群的任务再平衡过程，当 Worker 加入或离开时触发
- **ForceNew**: Terraform 中的属性标记，表示修改该字段会触发资源销毁重建

## Requirements

### Requirement 1: ConnectCluster 资源生命周期管理

**User Story:** As a platform operator, I want to create and manage Connect Worker clusters independently from Connectors, so that I can share one cluster across multiple Connectors to reduce resource overhead.

#### Acceptance Criteria

1. WHEN a user submits a ConnectCluster creation request with valid parameters, THE CMP SHALL create a new ConnectCluster resource with a unique ID (format: `connect-cluster-{uuid8}`), deploy a K8s Deployment with the specified capacity, and transition the state to STEADY upon successful deployment.
2. WHEN a user submits a ConnectCluster update request, THE CMP SHALL compute a change plan, apply the changes (rolling restart if Pod-level configuration changed, HPA adjustment if capacity changed), and transition the state back to STEADY upon completion.
3. WHEN a user submits a ConnectCluster deletion request and no Connector references the cluster, THE CMP SHALL delete the K8s resources, clean up Internal_Topic, and remove the database record.
4. IF a user submits a ConnectCluster deletion request while Connectors still reference the cluster, THEN THE CMP SHALL reject the deletion with an error message indicating the referenced Connector IDs.
5. WHEN a user queries a ConnectCluster by ID, THE CMP SHALL return the cluster metadata, current state, capacity information, plugin list, and Worker status.
6. WHEN a user queries the ConnectCluster list with optional filters (name, kafkaInstanceId, state), THE CMP SHALL return a paginated list of matching clusters.

### Requirement 2: ConnectCluster 插件集管理

**User Story:** As a platform operator, I want to declare a list of plugins at the cluster level, so that multiple Connectors using different plugins can share the same Worker cluster.

#### Acceptance Criteria

1. WHEN a user creates a ConnectCluster with a plugins list, THE CMP SHALL validate that no two entries share the same plugin name with different versions, and include all declared plugins in the Worker container image.
2. IF a user creates or updates a ConnectCluster with a plugins list containing the same plugin name at different versions, THEN THE CMP SHALL reject the request with a version conflict error.
3. WHEN a user updates the plugins list of an existing ConnectCluster, THE CMP SHALL trigger a Worker rolling restart to load the new plugin set, while Connector states remain unaffected.

### Requirement 3: ConnectCluster 容量管理

**User Story:** As a platform operator, I want to choose between provisioned (fixed) and autoscaling capacity modes for a Connect cluster, so that I can balance cost and elasticity based on workload characteristics.

#### Acceptance Criteria

1. WHEN a user creates a ConnectCluster with capacity type "provisioned", THE CMP SHALL deploy the specified number of Worker Pods with the specified resource spec (TIER1/TIER2/TIER3/TIER4).
2. WHEN a user creates a ConnectCluster with capacity type "autoscaling", THE CMP SHALL deploy a minimum number of Worker Pods and create an HPA resource targeting Pod CPU utilization, with min_worker_count as the lower bound and max_worker_count as the upper bound.
3. WHILE a ConnectCluster is in autoscaling mode and Pod CPU utilization exceeds the scale_out threshold, THE CMP SHALL allow HPA to add Worker Pods up to max_worker_count.
4. WHILE a ConnectCluster is in autoscaling mode and Pod CPU utilization drops below the scale_in threshold, THE CMP SHALL allow HPA to remove Worker Pods down to min_worker_count.

### Requirement 4: ConnectCluster Worker 配置强制策略

**User Story:** As a platform operator, I want the system to enforce critical Worker-level configurations automatically, so that cluster stability and security are guaranteed without manual intervention.

#### Acceptance Criteria

1. THE CMP SHALL inject `connect.protocol=compatible` into every ConnectCluster Worker configuration to enforce incremental cooperative Rebalance.
2. THE CMP SHALL inject `connector.client.config.override.policy=None` into every ConnectCluster Worker configuration to prevent Connectors from overriding the Kafka cluster connection.
3. WHEN a ConnectCluster is created, THE CMP SHALL create Internal_Topic with fixed partition counts: offset topic = 16 partitions, status topic = 16 partitions, config topic = 1 partition.

### Requirement 5: Connector 资源生命周期管理

**User Story:** As a data engineer, I want to create, update, and delete Connectors independently from the Worker cluster, so that Connector operations complete in seconds without waiting for infrastructure provisioning.

#### Acceptance Criteria

1. WHEN a user submits a Connector creation request with a valid connect_cluster_id, connector_class, and task_count, THE CMP SHALL create the Connector via Connect_REST_API `POST /connectors`, persist the record to the database, and return the Connector metadata.
2. WHEN a user submits a Connector creation request with initial_offsets, THE CMP SHALL first create the Connector, then apply the offsets via Connect_REST_API `PATCH /connectors/{name}/offsets`.
3. WHEN a user submits a Connector update request with changed connector_config or task_count, THE CMP SHALL update the Connector via Connect_REST_API `PUT /connectors/{name}/config` and update the database record.
4. WHEN a user submits a Connector deletion request, THE CMP SHALL delete the Connector via Connect_REST_API `DELETE /connectors/{name}` and remove the database record, without cleaning up offset data in Internal_Topic.
5. WHEN a user queries a Connector by ID, THE CMP SHALL return the Connector metadata from the database and the real-time runtime status from Connect_REST_API.
6. WHEN a user queries the Connector list with optional filters (name, connectClusterId, state), THE CMP SHALL return a paginated list of matching Connectors.

### Requirement 6: Connector 名称唯一性与状态独立性

**User Story:** As a data engineer, I want Connector names to be globally unique and Connector states to be independent from the cluster state, so that I can manage Connectors without confusion and plan for future cross-cluster migration.

#### Acceptance Criteria

1. WHEN a user creates a Connector, THE CMP SHALL validate that the Connector name is globally unique across all ConnectClusters, and reject the request with a duplicate name error if the name already exists.
2. WHILE a ConnectCluster is in ROLLING_UPDATE state, THE CMP SHALL report each Connector state independently based on its own runtime status from Connect_REST_API, without propagating the cluster state to Connectors.
3. WHEN a ConnectCluster transitions from ROLLING_UPDATE to STEADY, THE CMP SHALL not change any Connector state as a side effect.

### Requirement 7: Connector 在 Cluster 未就绪时的排队机制

**User Story:** As a data engineer, I want to create Connectors even when the target cluster is still provisioning, so that I can declare my full infrastructure in Terraform without worrying about resource ordering.

#### Acceptance Criteria

1. WHEN a user creates a Connector targeting a ConnectCluster that is not yet in STEADY state, THE CMP SHALL accept the creation request, persist the Connector record with a PENDING state, and enqueue the Connect_REST_API call.
2. WHEN the target ConnectCluster transitions to STEADY state, THE CMP SHALL automatically submit all enqueued Connector creation requests via Connect_REST_API and update their states accordingly.
3. IF the target ConnectCluster fails to reach STEADY state, THEN THE CMP SHALL transition enqueued Connectors to FAILED state with an error message referencing the cluster failure.

### Requirement 8: Connector 运行时控制

**User Story:** As a data engineer, I want to pause, resume, and restart individual Connectors without affecting other Connectors on the same cluster, so that I can perform maintenance on specific data pipelines.

#### Acceptance Criteria

1. WHEN a user submits a pause request for a Connector, THE CMP SHALL call Connect_REST_API `PUT /connectors/{name}/pause` and update the Connector state to PAUSED.
2. WHEN a user submits a resume request for a paused Connector, THE CMP SHALL call Connect_REST_API `PUT /connectors/{name}/resume` and update the Connector state based on the runtime status.
3. WHEN a user submits a restart request for a Connector, THE CMP SHALL call Connect_REST_API `POST /connectors/{name}/restart` and update the Connector state based on the runtime status.

### Requirement 9: 后端 API 端点设计

**User Story:** As a Terraform Provider developer, I want clearly separated REST API endpoints for ConnectCluster and Connector resources, so that I can implement two independent Terraform resources with clean CRUD semantics.

#### Acceptance Criteria

1. THE CMP SHALL expose ConnectCluster management endpoints under `/api/v1/connect-clusters` with standard CRUD operations (POST, GET, PUT, DELETE) plus Worker status (`GET /{id}/workers`), metrics (`GET /{id}/metrics`), logs (`GET /{id}/logs`), and version listing (`GET /versions`).
2. THE CMP SHALL expose Connector management endpoints under `/api/v1/connectors` with standard CRUD operations (POST, GET, PUT, DELETE) plus runtime control (`POST /{id}:pause`, `POST /{id}:resume`, `POST /{id}:restart`) and task listing (`GET /{id}/tasks`).
3. THE CMP SHALL support filtering Connectors by `connectClusterId` query parameter on the list endpoint (`GET /api/v1/connectors?connectClusterId=xxx`).

### Requirement 10: 数据库迁移与向后兼容

**User Story:** As a platform operator, I want existing single-tenant Connect deployments to be automatically migrated to the new two-layer model, so that the upgrade is seamless and no data is lost.

#### Acceptance Criteria

1. WHEN the CMP is upgraded to the multi-tenant version, THE CMP SHALL execute a database migration that creates the `cmp_connect_cluster` table and migrates cluster-related fields from each existing `cmp_connect_instance` record into a new `cmp_connect_cluster` record.
2. WHEN the database migration completes, THE CMP SHALL add a `connect_cluster_id` column to `cmp_connect_instance` and back-fill it with the corresponding cluster ID, resulting in a 1:1 mapping for all existing records.
3. WHEN the database migration completes, THE CMP SHALL preserve all existing Connector functionality (create, update, delete, pause, resume, restart, query) without requiring user intervention.

### Requirement 11: Terraform Provider — automq_connect_cluster 资源

**User Story:** As a DevOps engineer, I want a Terraform resource `automq_connect_cluster` to declaratively manage Connect Worker clusters, so that I can version-control my Connect infrastructure alongside other AutoMQ resources.

#### Acceptance Criteria

1. THE Terraform_Provider SHALL implement `automq_connect_cluster` resource with Required attributes: environment_id (ForceNew), name, plugins (list of {name, version}), kafka_cluster (containing kafka_instance_id (ForceNew) and security_protocol), capacity (containing type and provisioned or autoscaling block), and compute (containing type (ForceNew) and kubernetes block (ForceNew)).
2. THE Terraform_Provider SHALL implement `automq_connect_cluster` resource with Optional attributes: description, worker_config (map), metric_exporter, tags, version, and compute.iam_role (ForceNew).
3. THE Terraform_Provider SHALL implement `automq_connect_cluster` resource with Computed attributes: id, state, kafka_connect_version, created_at, updated_at.
4. WHEN a user runs `terraform apply` with an `automq_connect_cluster` resource, THE Terraform_Provider SHALL call the CMP ConnectCluster creation API and poll until the cluster reaches STEADY state or the create timeout expires.
5. WHEN a user runs `terraform destroy` on an `automq_connect_cluster` resource, THE Terraform_Provider SHALL call the CMP ConnectCluster deletion API and poll until the cluster is fully removed or the delete timeout expires.

### Requirement 12: Terraform Provider — automq_connector 资源重构

**User Story:** As a DevOps engineer, I want the `automq_connector` Terraform resource to reference a `connect_cluster_id` instead of embedding infrastructure parameters, so that Connector definitions are lightweight and focused on business configuration.

#### Acceptance Criteria

1. THE Terraform_Provider SHALL refactor `automq_connector` resource to require connect_cluster_id (ForceNew) and connector_class (ForceNew) as Required attributes, and remove infrastructure-related attributes (kubernetes_cluster_id, kubernetes_namespace, kubernetes_service_account, iam_role, capacity, worker_config, metric_exporter, version, scheduling_spec).
2. THE Terraform_Provider SHALL implement `automq_connector` resource with Required attributes: environment_id (ForceNew), connect_cluster_id (ForceNew), name, connector_class (ForceNew), task_count.
3. THE Terraform_Provider SHALL implement `automq_connector` resource with Optional attributes: description, connector_config (map), connector_config_sensitive (map, Sensitive), initial_offsets (list, Create-only).
4. THE Terraform_Provider SHALL implement `automq_connector` resource with Computed attributes: id, state, connector_type, plugin_id, created_at, updated_at.
5. WHEN a user runs `terraform apply` with an `automq_connector` resource, THE Terraform_Provider SHALL call the CMP Connector creation API and return immediately upon success (no polling required, as Connector creation is synchronous via Connect_REST_API).

### Requirement 13: 前端页面结构重构

**User Story:** As a platform operator using the CMP console, I want separate management pages for Connect clusters and Connectors, so that I can manage infrastructure and business logic independently.

#### Acceptance Criteria

1. THE CMP_Frontend SHALL provide a ConnectCluster list page at `/connect-clusters` displaying cluster name, state, Worker count, plugin list, and capacity type.
2. THE CMP_Frontend SHALL provide a ConnectCluster creation page at `/connect-clusters/create` with form steps: configure plugins → configure Kafka connection + compute + capacity → configure Worker settings → confirm.
3. THE CMP_Frontend SHALL provide a ConnectCluster detail page at `/connect-clusters/detail/:clusterId` with tabs: Overview, Connectors (list of Connectors on this cluster), Workers, Configuration, Logs, Metrics.
4. THE CMP_Frontend SHALL provide a Connector creation page at `/connect-clusters/:clusterId/connectors/create` with form steps: select connector_class → configure connector settings → confirm.
5. THE CMP_Frontend SHALL provide a Connector detail page at `/connectors/:connectorId` with tabs: Overview, Tasks, Configuration.

### Requirement 14: AIP 文档产出

**User Story:** As a project lead, I want a complete AIP document covering the multi-tenant refactoring, so that the design can be reviewed and approved before development begins.

#### Acceptance Criteria

1. THE AIP_Document SHALL follow the standard AIP template structure: 背景, 问题定义, 调研论证, 解决方案, 原型设计, 接口设计, 依赖选型, 方案详情, 兼容性问题, 被拒绝的其他方案, 落地计划.
2. THE AIP_Document SHALL include all 21 reviewed design decisions (Q1–Q21) with their rationale and implementation mapping.
3. THE AIP_Document SHALL include the complete API schema for both ConnectCluster and Connector resources (REST API and Terraform HCL).
4. THE AIP_Document SHALL include the database migration plan with SQL statements and backward compatibility analysis.
5. THE AIP_Document SHALL include a phased execution plan with estimated timelines for each phase (data model, manager, service, controller, frontend, testing).
