# Implementation Plan: CMP Connect 多租户重构

## Overview

将 CMP Connect 模块从单资源模型重构为两层资源模型（ConnectCluster + Connector），涉及后端（Java Spring Boot）、前端（React + Cloudscape）、Terraform Provider（Go）三个代码库。按 6 个阶段递进实施，后端优先，前端和 Terraform Provider 在后端 API 就绪后并行。

## Tasks

- [x] 1. AIP 文档产出
  - AIP 文档已完成并提交至 `docs/design/aip-connect-multi-tenant.md`
  - 包含 21 个设计决策、API Schema、数据库迁移方案、分阶段执行计划
  - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5_

- [x] 2. Phase 1: 数据模型 + 数据库迁移（cmp-common + cmp-repository）
  - [x] 2.1 新建 ConnectCluster Entity 及相关值对象
    - 创建 `model/entity/connect/ConnectCluster.java`，包含集群基本信息、插件集、Kafka 连接、K8s 部署、容量、Worker 配置、指标、运行时状态、元数据字段
    - 创建 `model/entity/connect/ClusterPlugin.java` 值对象（name + version）
    - 创建 `model/entity/connect/InitialOffset.java` 值对象（partition + offset map）
    - 创建 `ConnectClusterState` 枚举（CREATE_SUBMITTED, APPLYING_MANIFEST, WAITING_FOR_PODS, STEADY, ROLLING_UPDATE, DELETE_SUBMITTED, DELETING, FAILED）
    - _Requirements: 1.1, 1.2, 2.1, 3.1, 3.2_

  - [x] 2.2 精简 Connector Entity
    - 修改 `KubernetesConnector.java`（或新建 `Connector.java`），移除基础设施字段（kubernetes_*, capacity_*, worker_*, kafka_*, deployment_name 等）
    - 新增 `connectClusterId`、`connectorPropertiesSensitive`、`initialOffsets` 字段
    - 确保 `ConnectorState` 包含 PENDING 状态（Cluster 未就绪时排队用）
    - _Requirements: 5.1, 6.1, 7.1_

  - [x] 2.3 新建 ConnectCluster Param/VO/Query 类
    - 创建 `ConnectClusterCreateParam.java`（name, description, plugins, kafkaCluster, capacity, compute, workerConfig, metricExporter, tags, version）
    - 创建 `ConnectClusterUpdateParam.java`（name, description, plugins, capacity, workerConfig, securityProtocolConfig, metricExporter, tags, version, schedulingSpec）
    - 创建 `ConnectClusterVO.java`（完整响应字段 + 状态映射）
    - 创建 `ConnectClusterQuery.java`（name, kafkaInstanceId, state 过滤）
    - _Requirements: 1.5, 1.6, 9.1_

  - [x] 2.4 精简 Connector Param/VO/Query 类
    - 修改 `ConnectorCreateParam.java`：新增 connectClusterId、connectorConfigSensitive、initialOffsets，移除基础设施参数
    - 修改 `ConnectorUpdateParam.java`：精简为 name、description、taskCount、connectorConfig、connectorConfigSensitive
    - 修改 `ConnectorVO.java`：新增 connectClusterId，移除基础设施字段，新增 connectorType(Computed)、pluginId(Computed)
    - 修改 `ConnectorQuery.java`：新增 connectClusterId 过滤
    - _Requirements: 5.1, 5.6, 9.2, 9.3_

  - [x] 2.5 新建 ConnectCluster DAO 层
    - 创建 `cmp_connect_cluster` 建表 SQL（SQLite DDL）
    - 创建 `ConnectClusterDO.java` 数据对象
    - 创建 `ConnectClusterDAO.java` 接口（get, getByCode, query, count, insert, update, delete, updatePodEndpoints）
    - 创建 `ConnectClusterDAOImpl.java`（SQLite 实现）
    - 创建 MyBatis Mapper XML
    - 创建 `ConnectClusterConvertor.java`（DO ↔ Entity 转换）
    - _Requirements: 1.1, 1.5, 1.6_

  - [x] 2.6 修改 Connector DAO 层
    - 修改 `cmp_connect_instance` 表：新增 `connect_cluster_id`、`connector_config_sensitive` 列
    - 修改 `ConnectInstanceDO.java`：新增对应字段
    - 修改 Mapper XML：更新 insert/update/select 语句，新增 connectClusterId 过滤
    - 修改 `ConnectorConvertor.java`：适配新字段
    - _Requirements: 10.1, 10.2_

  - [x] 2.7 编写数据库迁移脚本
    - 创建 `cmp-app/src/main/resources/sql/v8.3.0_connect_multi_tenant.sql`
    - 所有语句必须幂等（CMP 使用 SimpleDDLExecutor，启动时自动执行所有 sql/*.sql，失败只 warn 不中断）
    - `CREATE TABLE IF NOT EXISTS cmp_connect_cluster`（新表，幂等）
    - `ALTER TABLE cmp_connect_instance ADD COLUMN connect_cluster_id/connector_config_sensitive`（列已存在时 SQLite 报错但被 catch warn，幂等）
    - `INSERT OR IGNORE INTO cmp_connect_cluster SELECT ... FROM cmp_connect_instance`（从旧表迁移集群数据，INSERT OR IGNORE 防重复）
    - `UPDATE cmp_connect_instance SET connect_cluster_id = 'cluster-' || instance_id WHERE connect_cluster_id IS NULL`（回填，WHERE 条件防重复）
    - BYOC 客户升级时 CMP 重启即自动执行，无需手动干预
    - 旧列保留不删除（SQLite 不支持 DROP COLUMN），旧代码回滚兼容
    - _Requirements: 10.1, 10.2, 10.3_

  - [ ]* 2.8 Write property test for database migration data integrity
    - **Property 9: Database migration preserves data integrity**
    - Generator: 随机生成 `cmp_connect_instance` 记录（含各种字段组合）
    - Assertion: 迁移后 cluster 记录存在、connector 的 connect_cluster_id 正确引用、字段无丢失
    - **Validates: Requirements 10.1, 10.2**

- [x] 3. Checkpoint — Phase 1 验证
  - Ensure all tests pass, ask the user if questions arise.
  - 验证建表 SQL 可执行、迁移脚本在现有数据上正确运行、DO ↔ Entity 转换无数据丢失

- [x] 4. Phase 2: Manager 层（cmp-service）
  - [x] 4.1 新建 ConnectClusterManager
    - 创建 `ConnectClusterManager.java` 接口（create, update, delete, getById）
    - 创建 `ConnectClusterManagerImpl.java`：DB 持久化 + 异步 Task 提交（创建/更新/删除）
    - 从 `KubernetesConnectorManagerImpl` 抽出集群管理逻辑
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 4.2 精简 ConnectorManager
    - 重构 `KubernetesConnectorManager` → `ConnectorManager`（或保留原名精简）
    - 移除 K8s 资源管理逻辑，只保留 DB 持久化（create, update, delete, getById）
    - 新增 `listByClusterId(String clusterId)` 方法（删除 Cluster 前校验用）
    - _Requirements: 5.1, 5.4, 1.4_

  - [x] 4.3 调整 ConnectClusterDeployManager
    - 确认 `buildConnectManifest()` 支持多插件列表
    - 新增 HPA YAML 生成逻辑（autoscaling 模式）
    - 确保 manifest 注入强制配置：`connect.protocol=compatible`、`connector.client.config.override.policy=None`、内部 Topic partition 数
    - _Requirements: 2.1, 3.2, 4.1, 4.2, 4.3_

  - [x] 4.4 调整异步 Task Steps
    - 重命名/调整 Step 类为 ConnectCluster 专属（ConnectClusterApplyStep, ConnectClusterCheckReadyStep, ConnectClusterInitEndpointsStep）
    - 移除 `BootstrapConnectorStep`（Connector 创建独立于 Cluster）
    - _Requirements: 1.1, 5.1_

- [x] 5. Phase 3: Service 层（cmp-service）
  - [x] 5.1 新建 ConnectClusterService
    - 创建 `ConnectClusterService.java` 接口
    - 创建 `ConnectClusterServiceImpl.java`，实现：
      - `create()`: 校验 plugins 版本冲突（Q2）、校验 K8s 可达性、强制注入 Worker 配置（Q9/Q18）、生成内部 Topic 名称和 partition 数（Q3）、持久化 + 提交异步 Task
      - `update()`: 计算变更计划（plugins 变化→滚动重启、capacity 变化→HPA/replica 调整）、计算 rollingId 指纹
      - `delete()`: 前置校验有无 Connector 引用（Q14）、清理内部 Topic（Q13）、提交异步删除 Task
      - `getById()` / `query()`: 标准查询 + 状态映射
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 2.1, 2.2, 2.3, 4.1, 4.2, 4.3_

  - [ ]* 5.2 Write property test for plugin version conflict validation
    - **Property 1: Plugin version conflict validation**
    - Generator: 随机 plugins 列表（0-10 个插件，随机 name/version）
    - Assertion: 同名不同版本 → 拒绝，否则 → 接受
    - **Validates: Requirements 2.1, 2.2**

  - [ ]* 5.3 Write property test for mandatory Worker config injection
    - **Property 2: Mandatory Worker config injection**
    - Generator: 随机 worker config map（可能包含/覆盖强制配置键）
    - Assertion: 结果始终包含 connect.protocol=compatible, connector.client.config.override.policy=None, offset.storage.partitions=16, status.storage.partitions=16, config.storage.partitions=1
    - **Validates: Requirements 4.1, 4.2, 4.3**

  - [ ]* 5.4 Write property test for cluster deletion guard
    - **Property 3: Cluster deletion guard**
    - Generator: 随机 cluster + 0-5 个 connectors
    - Assertion: connectors > 0 → 拒绝并返回 connector IDs，connectors == 0 → 接受
    - **Validates: Requirements 1.3, 1.4**

  - [x] 5.5 重构 ConnectorService
    - 重构 `ConnectorServiceImpl.java`，大幅简化：
      - `create()`: 校验 name 全局唯一（Q11）、校验 connectClusterId 存在、Cluster 未就绪则入队等待（Q10）、通过 ConnectRestClient `POST /connectors` 创建、如有 initialOffsets 调用 `PATCH /connectors/{name}/offsets`（Q12）、持久化
      - `update()`: 通过 ConnectRestClient `PUT /connectors/{name}/config` 更新
      - `delete()`: 通过 ConnectRestClient `DELETE /connectors/{name}` 删除，不清理 offset（Q13）
      - `getById()`: DB 读取 + ConnectRestClient 获取实时状态（Q15）
      - `pause/resume/restart`: 通过 ConnectRestClient 操作
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 6.1, 7.1, 8.1, 8.2, 8.3_

  - [ ]* 5.6 Write property test for Connector name global uniqueness
    - **Property 4: Connector name global uniqueness**
    - Generator: 两个 Connector 创建请求，随机 name（部分相同、部分不同）
    - Assertion: 同名 → 第二个被拒绝，不同名 → 都接受
    - **Validates: Requirements 6.1**

  - [ ]* 5.7 Write property test for Connector state independence
    - **Property 5: Connector state independence from cluster state**
    - Generator: 随机 cluster state × 随机 connector runtime status
    - Assertion: connector 报告状态仅由 runtime status 决定，不受 cluster state 影响
    - **Validates: Requirements 6.2, 6.3**

  - [ ]* 5.8 Write property test for Connector queuing
    - **Property 6: Connector queuing when cluster not STEADY**
    - Generator: 随机 cluster state（STEADY vs 非 STEADY）
    - Assertion: 非 STEADY → Connector 状态为 PENDING，STEADY → 直接提交到 Connect REST API
    - **Validates: Requirements 7.1**

  - [x] 5.9 新建 ConnectClusterAssembler + 调整 ConnectorAssembler
    - 创建 `ConnectClusterAssembler.java`：Entity → VO 转换，内部状态 → 外部状态映射
    - 调整 `ConnectorAssembler.java`：状态独立性（不传导 Cluster 状态）、新增 connectorType/pluginId 推导
    - _Requirements: 6.2, 6.3_

  - [x] 5.10 调整 ConnectorUpdatePlanner
    - 拆分为 `ConnectClusterUpdatePlanner`（plugins 变化→ROLLING_DEPLOY、capacity 变化→HPA/replica、worker config 变化→ROLLING_DEPLOY）+ `ConnectorUpdatePlanner`（connector config/task count 变化）
    - _Requirements: 1.2, 2.3_

  - [ ]* 5.11 Write property test for update change plan determinism
    - **Property 12: Update change plan determinism**
    - Generator: 随机 existing cluster + 随机 update params
    - Assertion: 相同输入 → 相同 action set，plugin 变化 → ROLLING_DEPLOY，rollingId 变化当且仅当 Pod 级配置变化
    - **Validates: Requirements 1.2**

  - [x] 5.12 实现 Connector 排队机制
    - 实现 Cluster 就绪后自动提交排队 Connector 的逻辑（在 ConnectClusterInitEndpointsStep 完成后触发）
    - 实现 Cluster FAILED 时将排队 Connector 转为 FAILED 的逻辑
    - _Requirements: 7.1, 7.2, 7.3_

- [x] 6. Checkpoint — Phase 2 + 3 验证
  - Ensure all tests pass, ask the user if questions arise.
  - 验证 ConnectClusterService 和 ConnectorService 核心逻辑正确，property tests 通过

- [x] 7. Phase 4: Controller 层（cmp-app）
  - [x] 7.1 新建 ConnectClusterController
    - 创建 `ConnectClusterController.java`，路径 `/api/v1/connect-clusters`
    - 实现端点：POST /（创建）、GET /{id}（查询）、PUT /{id}（更新）、DELETE /{id}（删除）、GET /（列表）、GET /{id}/workers、GET /{id}/metrics、GET /{id}/logs、GET /versions
    - 配置权限模型（CONNECT_CLUSTER_CREATE, CONNECT_CLUSTER_VIEW, CONNECT_CLUSTER_UPDATE, CONNECT_CLUSTER_DELETE, CONNECT_CLUSTER_LIST）
    - _Requirements: 9.1_

  - [x] 7.2 重构 ConnectorController
    - 修改 `ConnectorController.java`，精简参数
    - 确保端点：POST /（创建）、GET /{id}（查询）、PUT /{id}（更新）、DELETE /{id}（删除）、GET /（列表，支持 ?connectClusterId= 过滤）、POST /{id}:pause、POST /{id}:resume、POST /{id}:restart、GET /{id}/tasks
    - _Requirements: 9.2, 9.3_

  - [x] 7.3 新建/精简 Param/VO 的 JSON 序列化验证
    - 确保 ConnectClusterCreateParam、ConnectorCreateParam 的 @Valid 注解正确
    - 确保 VO 的 JSON 序列化字段名与 API 设计一致
    - _Requirements: 9.1, 9.2_

  - [ ]* 7.4 Write property test for ConnectCluster list filtering
    - **Property 7: ConnectCluster list filtering correctness**
    - Generator: 随机 ConnectCluster 集合 + 随机 filter criteria
    - Assertion: 返回列表中每个 cluster 都匹配所有 filter，且无遗漏
    - **Validates: Requirements 1.6**

  - [ ]* 7.5 Write property test for Connector list filtering
    - **Property 8: Connector list filtering correctness**
    - Generator: 随机 Connector 集合 + 随机 filter criteria
    - Assertion: 返回列表中每个 connector 都匹配所有 filter，且无遗漏
    - **Validates: Requirements 5.6**

- [x] 8. Checkpoint — Phase 4 验证
  - Ensure all tests pass, ask the user if questions arise.
  - 验证 API 端点可访问、参数校验正确、权限控制生效

- [x] 9. Phase 5: 前端改造（cmp-frontend-next）
  - [x] 9.1 API 层调整
    - 修改 `service/connect.ts`，新增 ConnectCluster API（getConnectClusters, createConnectCluster, getConnectCluster, updateConnectCluster, deleteConnectCluster, listClusterWorkers, getClusterMetrics, getClusterLogs）
    - 新增 Connector API（getConnectors, createConnector, getConnector, updateConnector, deleteConnector, pauseConnector, resumeConnector, restartConnector, listConnectorTasks）
    - 定义 TypeScript 类型（ConnectCluster, Connector, ConnectClusterCreateParam 等）
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5_

  - [x] 9.2 ConnectCluster 列表页改造
    - 改造 `/connect-clusters` 页面，展示集群 name、state、Worker count、plugin list、capacity type
    - 支持创建/删除操作入口
    - _Requirements: 13.1_

  - [x] 9.3 ConnectCluster 创建页改造
    - 改造 `/connect-clusters/create` 页面
    - 表单步骤：配置 plugins → 配置 Kafka 连接 + compute + capacity → 配置 Worker 设置 → 确认
    - 移除原有的 Connector 配置步骤，新增 plugins 多选配置
    - _Requirements: 13.2_

  - [x] 9.4 ConnectCluster 详情页改造
    - 改造 `/connect-clusters/detail/:clusterId` 页面
    - Tabs: Overview, Connectors（该集群下的 Connector 列表）, Workers, Configuration, Logs, Metrics
    - 新增 Connectors Tab 展示该集群下所有 Connector
    - 保留 resize、update-config、update-metrics-exporter、upgrade-version 操作
    - 新增 update-plugins 操作页面
    - _Requirements: 13.3_

  - [x] 9.5 新增 Connector 创建页
    - 创建 `/connect-clusters/:clusterId/connectors/create` 页面
    - 表单步骤：选择 connector_class → 配置 connector settings（config + sensitive config + task count）→ 确认
    - _Requirements: 13.4_

  - [x] 9.6 新增 Connector 详情页
    - 创建 `/connectors/:connectorId` 页面
    - Tabs: Overview, Tasks, Configuration
    - 支持 pause/resume/restart 操作按钮
    - 新增 update-config 操作页面
    - _Requirements: 13.5_

- [x] 10. Checkpoint — Phase 5 验证
  - Ensure all tests pass, ask the user if questions arise.
  - 验证前端页面导航正确、表单提交正常、状态展示准确

- [ ] 11. Phase 6: Terraform Provider 适配 + 集成测试
  - [ ] 11.1 新建 ConnectCluster API client
    - 创建 `client/api_connect_cluster.go`：CRUD 方法（CreateConnectCluster, GetConnectCluster, UpdateConnectCluster, DeleteConnectCluster, ListConnectClusters）
    - 创建 `client/model_connect_cluster.go`：请求参数（ConnectClusterCreateParam 等）+ 响应 VO（ConnectClusterVO 等）+ 状态常量
    - _Requirements: 11.1, 11.2, 11.3_

  - [ ] 11.2 新建 ConnectCluster Expand/Flatten 转换层
    - 创建 `internal/models/connect_cluster.go`：ConnectClusterResourceModel 结构体 + ExpandConnectClusterCreate + FlattenConnectCluster
    - 处理嵌套结构：plugins list、kafka_cluster block、capacity block（provisioned/autoscaling）、compute block（kubernetes）
    - _Requirements: 11.1, 11.2, 11.3_

  - [ ]* 11.3 Write property test for ConnectCluster Expand/Flatten round-trip
    - **Property 10: Terraform ConnectCluster Expand/Flatten round-trip**
    - Library: `pgregory.net/rapid`
    - Generator: 随机 ConnectClusterResourceModel（valid plans）
    - Assertion: Expand → mock API response → Flatten ≈ original plan（non-Computed fields）
    - **Validates: Requirements 11.1, 11.2, 11.3**

  - [ ] 11.4 新建 automq_connect_cluster resource
    - 创建 `internal/provider/resource_connect_cluster.go`：Schema 定义 + Create/Read/Update/Delete/ImportState handler
    - Create: 调用 CMP API → 轮询等待 STEADY 状态（timeout 30m）
    - Delete: 调用 CMP API → 轮询等待删除完成（timeout 20m）
    - Read: 404 → RemoveResource
    - ImportState: 解析 `{environment_id}@{cluster_id}` 格式
    - Schema Required: environment_id(ForceNew), name, plugins, kafka_cluster, capacity, compute
    - Schema Optional: description, worker_config, metric_exporter, tags, version
    - Schema Computed: id, state, kafka_connect_version, created_at, updated_at
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

  - [ ] 11.5 重构 automq_connector resource
    - 修改 `client/model_connector.go`：精简请求参数，新增 connect_cluster_id、connector_config_sensitive、initial_offsets
    - 重写 `internal/models/connector.go`：移除基础设施字段，新增 connect_cluster_id、connector_class(ForceNew)、connector_config_sensitive(Sensitive)、initial_offsets(Create-only)
    - 重写 `internal/provider/resource_connector.go`：精简 Schema，移除基础设施属性，Create 同步返回（无需轮询）
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

  - [ ]* 11.6 Write property test for Connector Expand/Flatten round-trip
    - **Property 11: Terraform Connector Expand/Flatten round-trip**
    - Library: `pgregory.net/rapid`
    - Generator: 随机 ConnectorResourceModel（valid plans，含 connect_cluster_id、connector_class、name、task_count、optional connector_config）
    - Assertion: Expand → mock API response → Flatten ≈ original plan（non-Computed fields）
    - **Validates: Requirements 12.2, 12.3, 12.4**

  - [ ]* 11.7 Write unit tests for ConnectCluster and Connector Expand/Flatten
    - 测试边界条件：nil 字段、空 map、空 plugins 列表、sensitive 字段保留
    - 测试 cRetainStr 模式对 connector_config_sensitive 的处理
    - _Requirements: 11.1, 12.2_

  - [ ]* 11.8 Write backend integration tests
    - 创建 ConnectCluster → 验证 K8s Deployment 创建
    - 创建 ConnectCluster（autoscaling）→ 验证 HPA 创建
    - 创建 ConnectCluster → 创建多个 Connector → 验证多租共享
    - 更新 ConnectCluster plugins → 验证滚动重启 + Connector 状态不变
    - 删除 Connector → 验证 offset 未清理
    - 删除 ConnectCluster → 验证内部 Topic 清理
    - 数据迁移测试：现有 1:1 数据迁移后功能正常
    - _Requirements: 1.1, 1.3, 2.3, 5.1, 5.4, 10.1, 10.2, 10.3_

  - [ ] 11.9 更新 Terraform Provider 文档和示例
    - 运行 `go generate ./...` 重新生成文档
    - 更新 `examples/resources/automq_connect_cluster/` 示例
    - 更新 `examples/resources/automq_connector/` 示例（精简后）
    - 更新 `docs/resources/connector.md`
    - 新建 `docs/resources/connect_cluster.md`
    - _Requirements: 11.1, 12.1_

- [ ] 12. Final checkpoint — 全量验证
  - Ensure all tests pass, ask the user if questions arise.
  - 验证后端 API、前端页面、Terraform Provider 三端功能完整
  - 验证数据迁移在现有环境上正确执行
  - 验证多租户场景：1 个 ConnectCluster 运行 N 个 Connector

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Phase 1-4（后端）必须按顺序执行，Phase 5（前端）可在 Phase 4 完成后并行
- Phase 6 Terraform Provider 依赖后端 API 就绪（Phase 4 完成后）
- 后端代码在 `automqbox/` symlink 目录，分支 `feat/connect-multi-tenant`（从 `develop/202603` 切出）
- Property tests 使用 Java `jqwik` 库（后端）和 Go `rapid` 库（Terraform Provider）
- 每个 property test 最少运行 100 次迭代
