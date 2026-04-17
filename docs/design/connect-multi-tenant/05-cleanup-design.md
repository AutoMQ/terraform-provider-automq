# Connect 多租户重构 — Phase D: 状态分离 + 遗留代码清理设计

> 分支: feat/connect-multi-tenant on AutoMQ/automqbox
> 基准 commit: 80a6d0a796
> 日期: 2026-04-17
> 前置文档: 00-archaeology.md, 01-new-design.md, 02-diff-analysis.md

---

## 1. 问题总结

当前代码是"四不像"状态：多租户架构已搭建（ConnectCluster 独立实体 + Connector 精简实体），但大量遗留单租户代码未清理：

1. **Connector.state 类型错误**: `Connector` 实体的 `state` 字段类型是 `ConnectClusterState`，但多租户 Connector 的状态应该独立于 Cluster，使用 `ConnectorState`
2. **Update Task Steps 操作错误实体**: 5 个 `UpdateConnectCluster*Step` 通过 `ConnectorManager.get()` 读取 `Connector` 对象来操作 Cluster，而不是通过 `ConnectClusterManager.get()` 读取 `ConnectCluster`
3. **ConnectClusterAdapter 语义混乱**: 将 `ConnectCluster` 转为 `Connector` 的 adapter 是过渡期 workaround，但 `Connector` 实体上仍保留大量基础设施字段（K8s、容量、Worker 配置），导致两个实体字段高度重叠
4. **ConnectorManager 接口臃肿**: 包含 `updateStatus(ConnectorState)` 和 `updatePodEndpoints()` 等方法，这些是 Cluster 级别操作，不应在 Connector Manager 中
5. **ConnectorManagerImpl 保留完整旧模式代码**: K8s 清理、Kafka Topic 清理、Task 提交等逻辑仍在，通过 `connectClusterId != null` 分支区分
6. **ConnectorStateResolver 有两个 resolve 重载**: 一个接受 `ConnectorState`（多租户），一个接受 `ConnectClusterState`（旧模式），旧模式重载应删除
7. **ConnectorQuery.clusterState 字段**: 用于将 `ConnectorState` 映射为 `ConnectClusterState` 列表做 DB 查询，这是因为 DB 中存的是 `ConnectClusterState` 字符串

---

## 2. 设计目标

- Connector 实体的 `state` 字段改为 `ConnectorState` 类型
- Update Task Steps 改为操作 `ConnectCluster`（通过 `ConnectClusterManager`）
- 清理 `Connector` 实体上的基础设施字段（保留 DB 列但 Entity 不再使用）
- 清理 `ConnectorManagerImpl` 中的旧模式代码路径
- 清理 `ConnectorStateResolver` 的旧模式重载
- 清理 `ConnectorQuery.clusterState` 字段
- 所有变更必须编译通过

---

## 3. 详细设计

### 3.1 Connector.state: ConnectClusterState → ConnectorState

**变更**:
```java
// Before (Connector.java)
private ConnectClusterState state;

// After
private ConnectorState state;
```

**影响范围**:
- `ConnectorConvertorImpl.toEntity()` — 从 DO 的 `state` 字符串解析为 `ConnectorState`
- `ConnectorConvertorImpl.toDO()` — 将 `ConnectorState` 序列化为字符串
- `ConnectorAssemblerImpl` — 使用 `ConnectorState` 调用 `ConnectorStateResolver.resolve()`
- `ConnectorServiceImpl` — 所有 `setState()` 调用改为 `ConnectorState` 值
- `ConnectorQueueHandler` — `setState()` 调用改为 `ConnectorState` 值
- `ConnectClusterAdapter.toConnector()` — 需要 `mapClusterState()` 将 `ConnectClusterState` 映射为 `ConnectorState`

**DB 兼容性**: `cmp_connect_instance.state` 列存储的是字符串。旧数据存的是 `ConnectClusterState` 值（如 `STEADY`, `CREATE_SUBMITTED`）。`ConnectorConvertorImpl.toEntity()` 需要处理两种格式：
```java
private ConnectorState parseState(String stateStr) {
    if (stateStr == null) return ConnectorState.UNKNOWN;
    // 先尝试直接解析为 ConnectorState
    try {
        return ConnectorState.valueOf(stateStr);
    } catch (IllegalArgumentException e) {
        // 旧数据: ConnectClusterState 值，需要映射
        try {
            ConnectClusterState clusterState = ConnectClusterState.valueOf(stateStr);
            return mapClusterStateToConnectorState(clusterState);
        } catch (IllegalArgumentException e2) {
            return ConnectorState.UNKNOWN;
        }
    }
}
```

**映射规则**:
| ConnectClusterState | ConnectorState |
|---|---|
| CREATE_SUBMITTED, APPLYING_MANIFEST, WAITING_FOR_PODS, BOOTSTRAPPING_CONNECTOR | CREATING |
| STEADY | RUNNING |
| UPDATE_SUBMITTED, ROLLING_UPDATE, SYNCING_CONNECTOR, UPGRADING | CHANGING |
| DELETE_SUBMITTED, DELETING | DELETING |
| FAILED | FAILED |

### 3.2 Update Task Steps: ConnectorManager → ConnectClusterManager

**当前问题**: 5 个 Update Step 都注入 `ConnectorManager`，调用 `connectorManager.get(instanceId)` 获取 `Connector` 对象，然后操作 `Connector` 的基础设施字段。

**变更**: 改为注入 `ConnectClusterManager`，调用 `connectClusterManager.get(clusterId)` 获取 `ConnectCluster` 对象。

**受影响文件**:

1. **UpdateConnectClusterPrepareStep**
   - `ConnectorManager` → `ConnectClusterManager`
   - `connectInstanceManager.get(payload.getConnectInstanceId())` → `connectClusterManager.get(payload.getConnectInstanceId())`
   - 返回 `ConnectCluster` 而非 `Connector`

2. **UpdateConnectClusterApplyStep**
   - `ConnectorManager` → `ConnectClusterManager`
   - `connectInstanceManager.get()` → `connectClusterManager.get()`
   - `connectInstance.setState(ConnectorState.CHANGING)` → `connectClusterManager.updateStatus(id, ConnectClusterState.ROLLING_UPDATE)`
   - `connectInstanceManager.update(connectInstance)` → 通过 `connectClusterManager` 更新
   - 构建 manifest 时直接使用 `ConnectCluster`，不再需要 adapter 转换
   - `connectInstance.setWorkerCount()` / `setPluginId()` / `setProperties()` → 操作 `ConnectCluster` 对象

3. **UpdateConnectClusterCheckReadyStep**
   - `ConnectorManager` → `ConnectClusterManager`
   - `connectInstanceManager.get()` → `connectClusterManager.get()`
   - 从 `ConnectCluster` 读取 `deploymentName` 和 `kubernetesNamespace`

4. **UpdateConnectClusterEndpointStep**
   - `ConnectorManager` → `ConnectClusterManager`
   - `connectInstanceManager.get()` → `connectClusterManager.get()`
   - `deployManager.updatePodEndpoints(connectInstance, credentials)` → 需要 adapter 或直接传 `ConnectCluster`

5. **UpdateConnectClusterConnectorStep**
   - `ConnectorManager` → `ConnectClusterManager`
   - 这个 Step 操作的是 Cluster 下的 Connector 配置同步
   - `connectInstanceManager.get()` → `connectClusterManager.get()`
   - `connectInstance.setState(ConnectorState.RUNNING/CHANGING/FAILED)` → `connectClusterManager.updateStatus(id, ConnectClusterState.STEADY/ROLLING_UPDATE/FAILED)`
   - REST API 调用仍需要 `podEndPoints`，从 `ConnectCluster` 获取

**UpdateConnectTaskInstanceFactoryImpl** 也需要更新：构造 Step 时传入 `ConnectClusterManager` 而非 `ConnectorManager`。

### 3.3 ConnectClusterAdapter 简化

**当前**: `toConnector()` 将 `ConnectCluster` 的所有字段复制到 `Connector`。

**变更**: 
- Create Task Steps（`ConnectClusterApplyStep`, `ConnectClusterCheckReadyStep`, `ConnectClusterInitEndpointsStep`）已经通过 `ConnectClusterManager` 操作 `ConnectCluster`，但 `ConnectClusterDeployManager` 的接口仍接受 `Connector` 类型。
- 保留 `ConnectClusterAdapter.toConnector()` 用于 `ConnectClusterDeployManager` 接口兼容，但添加 `mapClusterState()` 方法处理状态映射。
- 长期目标：让 `ConnectClusterDeployManager` 直接接受 `ConnectCluster`，消除 adapter。本次不做。

### 3.4 ConnectorManager 接口清理

**删除方法**:
- `updatePodEndpoints(String instanceId, List<String> endpoints)` — 这是 Cluster 级别操作，已在 `ConnectClusterManager` 中
- `updateStatus(String instanceId, ConnectorState status)` — 多租户 Connector 的状态更新应通过 Service 层，不需要 Manager 暴露

**保留方法**:
- `create()`, `update()`, `delete()`, `get()`, `query()`, `listByClusterId()`

**实际操作**: 由于 `ConnectorManagerImpl` 内部的 `delete()` 旧模式路径仍调用 `updateStatus()`，将 `updateStatus()` 改为 `private` 方法而非接口方法。`updatePodEndpoints()` 在旧模式中由 `ConnectClusterDeployManagerImpl` 调用，检查是否仍需要。

### 3.5 ConnectorManagerImpl 清理

**删除/简化**:
- `create()` 中的 Task 提交逻辑 — 多租户 Connector 不需要 Task。添加 `connectClusterId != null` 判断，有 `connectClusterId` 时只做 DB 插入 + SecurityProtocolConfig 处理
- `cleanupKubernetesResources()` — 只在旧模式路径使用，保留但标记为 legacy
- `cleanupKafkaInternalTopics()` — 同上
- `setKubeConfig()` — 只在旧模式 `create()` 中使用，保留但标记为 legacy

### 3.6 ConnectorStateResolver 清理

**删除**: `resolve(ConnectorVO, ConnectClusterState)` 重载 — 旧模式不再需要，因为 `Connector.state` 已改为 `ConnectorState`

**保留**: `resolve(ConnectorVO, ConnectorState)` — 多租户 Connector 使用

**删除**: `map2ClusterState()` 静态方法 — 不再需要将 `ConnectorState` 映射为 `ConnectClusterState` 列表

### 3.7 ConnectorQuery.clusterState 清理

**删除**: `clusterState` 字段 — DB 查询直接使用 `state` 字段（现在存的是 `ConnectorState` 值）

**影响**: `ConnectorManagerImpl.query()` 中的 `query.setClusterState(ConnectorStateResolver.map2ClusterState(query.getState()))` 调用删除。MyBatis Mapper XML 中的 `clusterState` 条件需要改为直接使用 `state`。

### 3.8 前端 Bug 修复（从 043677dceb 中 cherry-pick 的修复）

以下前端修复在回滚时丢失了，需要重新实现：
1. Connector delete 确认使用 Cluster i18n keys → 改为 connector 专用 keys
2. Connector detail overview 描述说 "Connect Cluster" → 改为 "Connector"
3. Task 列表列头硬编码英文 → 国际化
4. `ConnectClusterServiceImpl.toVO()` 缺少 `workerResourceSpec` 映射 → 补充
5. update-config 页面提交后导航到 list → 改为导航到 detail
6. Cluster details-tab 显示 Connector 级别 Topics 字段 → 移除

---

## 4. 实施顺序

| 步骤 | 内容 | 依赖 |
|------|------|------|
| 1 | `Connector.java`: state 类型改为 `ConnectorState` | 无 |
| 2 | `ConnectorConvertorImpl`: 添加 `parseState()` 兼容旧数据 | 步骤 1 |
| 3 | `ConnectClusterAdapter`: 添加 `mapClusterState()` | 步骤 1 |
| 4 | `ConnectorServiceImpl`: 所有 `setState()` 改为 `ConnectorState` 值 | 步骤 1 |
| 5 | `ConnectorQueueHandler`: `setState()` 改为 `ConnectorState` 值 | 步骤 1 |
| 6 | `ConnectorAssemblerImpl`: 调用 `resolve(vo, ConnectorState)` | 步骤 1 |
| 7 | Update Task Steps: 改为注入 `ConnectClusterManager` | 步骤 1 |
| 8 | `UpdateConnectTaskInstanceFactoryImpl`: 传入 `ConnectClusterManager` | 步骤 7 |
| 9 | `ConnectorStateResolver`: 删除旧模式重载和 `map2ClusterState()` | 步骤 6 |
| 10 | `ConnectorQuery`: 删除 `clusterState` 字段 | 步骤 9 |
| 11 | `ConnectorManagerImpl.query()`: 删除 `clusterState` 映射 | 步骤 10 |
| 12 | MyBatis Mapper XML: `clusterState` 条件改为 `state` | 步骤 10 |
| 13 | `ConnectorManager` 接口: 清理不需要的方法 | 步骤 7 |
| 14 | `ConnectorManagerImpl`: 简化 `create()` 多租户路径 | 步骤 13 |
| 15 | 前端 Bug 修复 | 无依赖 |
| 16 | 编译验证 | 全部 |

---

## 5. 风险评估

| 风险 | 影响 | 缓解 |
|------|------|------|
| 旧数据 state 字符串不兼容 | 旧 Connector 的 state 是 `STEADY`/`CREATE_SUBMITTED` 等 | `parseState()` 双重解析兼容 |
| MyBatis Mapper XML 查询条件变更 | 查询可能返回错误结果 | 仔细检查 XML 中所有 `clusterState` 引用 |
| Update Task Steps 改动大 | 可能引入新 bug | 逐个 Step 改动，每改一个编译验证 |
| `ConnectClusterDeployManager` 接口仍接受 `Connector` | adapter 仍需保留 | 本次不改 DeployManager 接口，保留 adapter |
