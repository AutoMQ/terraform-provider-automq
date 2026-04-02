
# 插件沉淀标准工作流（SOP）

当用户说"开始新插件的沉淀"、"沉淀下一个插件"、"沉淀 XXX 插件"时，按本文档执行。

**启动前检查清单 — 在开始任何工作之前，先列出以下清单提醒用户准备，等用户确认全部就绪后再开始：**

1. CMP 环境地址、admin 用户名密码（用于创建 Service Account）
2. 环境 ID
3. kubectl 已配置好 EKS context（`kubectl get nodes` 能正常返回）— 如果 token 过期需要重新 `aws eks update-kubeconfig`
4. EKS 集群已添加 CMP 的访问条目（否则 connector 创建会报 KubernetesValidationFailed）
5. 如果插件需要外部 SaaS 资源（如 Snowflake 账号），提前准备好连接信息
6. S3 bucket 和 IAM Role（如果测试 S3 Sink 类插件）

**用户确认以上全部就绪后，我再开始执行。不要在用户没确认的情况下开始。**

**关键原则：执行过程中绝对不能中断。遇到任何问题都要自己想办法解决，不要停下来问用户。通过调 API、读源码、搜索文档等方式自行排障。只有在真正无法绕过的硬依赖（如需要用户提供 SaaS 账号密码）时才可以询问用户。**

**关键原则：遇到问题时必须先搜索源码再做判断，绝不能因为第一眼没找到就放弃。CMP 是一个完整的平台，大部分功能都有 API 支持。**

**关键原则：不要因为输出长度限制而跳过步骤或草草收尾。每个 SOP 步骤都必须完整执行，产出必须达到文档中定义的标准。如果快到输出限制，在回复末尾写"继续执行中..."让用户发"继续"触发下一轮，而不是把未完成的工作标记为"待完善"然后提交。**

**关键原则：性能测试不是跑一次就完了。SOP 要求的是多配置组合对比、多 Tier 对比、关系曲线图、CMP 指标截图。这些数据是给用户看的文档素材，必须完整。**

**关键原则：性能测试数据必须干净可用。如果测试过程中遇到网络错误或其他干扰导致数据不可靠，必须重跑直到拿到干净数据。文档是要放官网的，不能包含 network error 或异常数据。遇到网络问题时：1) 先排查原因（代理、超时配置、CMP 负载）2) 修复后重跑 3) 只有零错误的数据才能写入文档。**

**关键原则：文档必须包含性能指标的图片/图表。纯文字表格不够，需要可视化图表（throughput 曲线、Tier 对比柱状图、CMP 指标面板截图）。图表生成方案：用 Python matplotlib（已验证可用，需 `pip3 install --break-system-packages matplotlib`）。**

**关键原则：性能测试的目的和方法论。**
性能测试不是为了找到"最佳配置"，而是为了：
1. 给客户一个基准吞吐的认知 — 在特定条件下（Tier、Worker 数、Task 数），connector 能达到什么量级的吞吐
2. 展示关键参数对性能的影响趋势 — 比如 flush.size 从小到大，吞吐/延迟/文件大小如何变化；Tier 升级后吞吐提升多少
3. 给客户参数调优的方向指导 — 不是告诉客户"用这个配置"，而是让客户理解"这个参数影响什么，调大/调小会怎样"

因为实际生产环境有很多其他因素（消息大小、partition 数、下游 S3 延迟、网络带宽、数据格式等），我们只能在特定条件下测试出参数对吞吐的影响。文档中应该：
- 明确标注测试条件（消息大小、partition 数、格式等）
- 展示参数变化的趋势（不是绝对值）
- 说明"实际吞吐取决于你的具体场景"
- 给出不同场景下的推荐起点配置（低延迟/平衡/高吞吐），让客户在此基础上调优

测试应该覆盖的维度：
- flush.size 变化对吞吐和文件大小的影响
- Worker Tier 变化对吞吐的影响
- task_count 变化对吞吐的影响（需要足够的 partition）
- 压缩（gzip vs none）对吞吐和存储的影响
- 消息大小对吞吐的影响

当前测试的局限：通过 CMP Produce API 发送消息，吞吐受限于 API 的 HTTP 开销（~90 msgs/sec），无法测出 connector 的真实处理上限。要测真实上限需要用 Kafka 原生 producer 直接写入 topic，然后观察 connector 的消费速率。这个在 SOP 中需要改进。

**性能测试改进方案：用 Kafka Producer Client 直接压测**

CMP Produce API 吞吐上限约 90 msgs/sec（HTTP 签名开销），无法测出 connector 真实处理能力。正确做法：

1. 创建 instance 时使用**公网子网**，这样本地能直接连接 Kafka broker
   - 公网子网命名规则：`*-public-*`（如 `vpc-f7y28-automqlab-public-ap-southeast-1a`）
   - 私网子网命名规则：`*-private-*`
   - 当前环境公网子网：
     - ap-southeast-1a: `subnet-02028468e2f2d8b53`
     - ap-southeast-1b: `subnet-0bc4588f789d1eba6`
     - ap-southeast-1c: `subnet-0521140f0f90db879`

2. 用 Go 写一个 Kafka Producer 压测工具（使用 `github.com/IBM/sarama` 或 `github.com/confluentinc/confluent-kafka-go`）
   - 直接连接 Kafka broker（bootstrap servers 从 instance 详情的 endpoints 获取）
   - SASL_PLAINTEXT 认证（username/password）
   - 可配置：消息大小、发送速率、总消息数、并发度
   - 输出：produce throughput、P99 latency

3. 压测流程：
   - 用 producer 工具持续写入 topic（如 10000 msgs/sec × 5 分钟）
   - 同时观察 connector 的消费速率（通过 CMP metrics API 或 consumer lag）
   - 记录 connector 在不同 flush.size/Tier 下的实际消费吞吐

4. 这样能测出：
   - connector 的真实处理上限（不受 produce API 限制）
   - flush.size 对 S3 写入频率和延迟的真实影响
   - Tier 升级对处理能力的真实提升

需要创建的工具：`test-connector/kafka_producer_bench.go`（Kafka 原生 producer 压测工具）

**实际可行的压测方案：通过 kubectl 在 K8s 集群内运行 producer**

由于 Kafka broker 只有私网 endpoint（`.automq.private` DNS），从本地无法直连。正确做法是通过 kubectl 在同一个 K8s 集群内运行 producer pod：

1. **前提条件**：用户需要提供本地 kubectl 的 EKS context（`aws eks update-kubeconfig`）
   - 如果用户没有提供 kubectl context，**必须主动要求**
   - 验证：`kubectl config current-context` 应该返回对应的 EKS cluster ARN
   - 验证：`kubectl get nodes` 应该能列出节点

2. **压测方法**：用 `kubectl run` 启动临时 pod 运行 `kafka-console-producer` 或 `kafka-producer-perf-test`
   ```bash
   # 使用 AutoMQ 的 Kafka 镜像（包含 kafka-producer-perf-test.sh）
   kubectl run kafka-bench --rm -it --image=apache/kafka:3.9.0 -- \
     /opt/kafka/bin/kafka-producer-perf-test.sh \
     --topic s3-bench-topic \
     --num-records 50000 \
     --record-size 200 \
     --throughput -1 \
     --producer-props bootstrap.servers=0-{instanceId}.{domain}:9102 \
     security.protocol=SASL_PLAINTEXT \
     sasl.mechanism=SCRAM-SHA-512 \
     sasl.jaas.config='org.apache.kafka.common.security.scram.ScramLoginModule required username="connect-user" password="ConnectPoC123!";'
   ```

3. **为什么不能从本地直连**：
   - Kafka broker 的 advertised.listeners 用私网 IP/hostname
   - `.automq.private` DNS zone 只在 VPC 内部可解析
   - 即使用公网子网，EC2 的 Kafka listener 绑定私网 IP

**关键原则：使用 Spec 模式执行插件沉淀 SOP。在开始沉淀前，先创建一个 spec（requirements → design → tasks），把 SOP 的每个步骤和子步骤拆成 tasks.md 中的 checklist。这样即使被截断或跨多轮对话，也能从 tasks.md 中找到下一个未完成的 task 继续执行，不会跳步。**

## 自动化执行约定

### 1. 插件优先级与选型

- 排序依据：Confluent Hub / Confluent Cloud 的使用排名（目标是承接 Confluent 客户，用这个排名合理）
- 自动确定下一个：读取 `automqbox/cmp/cmp-service/src/main/resources/system-plugins.json` 查看已有内置插件，对照优先级列表找到下一个未沉淀的插件
- 不需要用户指定，自主决策
- 如果用户指定了具体插件名，直接沉淀该插件

### 2. 插件包存储位置

- 内置插件注册文件：`automqbox/cmp/cmp-service/src/main/resources/system-plugins.json`
- CDN 基础 URL：`http://download.automq.com/resource/connector/`
- 镜像构建配置：`automqbox/cmp/cmp-packer/image/vars.pkr.hcl`（plugins_cdn_base_url）
- Docker 构建：`automqbox/cmp/cmp-packer/docker/Dockerfile`
- 插件 CDN 文档：`automqbox/cmp/cmp-packer/docker/PLUGINS_CDN_CONFIG.md`

当前已有内置插件：

| 插件 | 版本 | 类型 | CDN URL |
|------|------|------|---------|
| debezium-postgresql | 3.1.2 | SOURCE | debezium-debezium-connector-postgresql-3.1.2.zip |
| debezium-mysql | 3.1.2 | SOURCE | debezium-debezium-connector-mysql-3.1.2.zip |
| aerospike-inbound | 3.2.0 | SINK | aerospike-kafka-inbound-3.2.0.zip |
| mongodb-kafka | 2.0.1 | SINK | mongodb-kafka-connect-mongodb-2.0.1.zip |
| clickhouse-sink | 1.3.4 | SINK | clickhouse-kafka-connect-v1.3.4.zip |
| clickzetta-connector | 1.0.0 | SINK | clickzetta-kafka-connector-1.0.1-connector.zip |
| snowflake | 3.4.0 | SINK | snowflake-3.4.0.zip |

注意：S3 Sink 插件（automq-kafka-connect-s3-11.1.0.zip）在 CDN 上存在但未注册到 system-plugins.json。

### 3. 测试环境

- 使用 CMP 环境，用户在让我做插件沉淀时会提供：CMP 地址、用户名密码、环境 ID
- **如果用户没有提供这些信息，我必须主动询问**
- **还需要 kubectl 的 EKS context**（用于在 K8s 集群内运行 Kafka producer 压测）。如果用户没有提供，必须主动要求
- **kubectl 认证会过期**（AWS STS token 有效期通常 1 小时）。如果 kubectl 报 "must be logged in" 错误，需要用户重新运行 `aws eks update-kubeconfig`
- 不要使用其他 steering 文件中记录的测试环境信息（可能已过期）

#### 环境初始化快速流程（已验证）

新 CMP 环境拿到后，按以下步骤初始化（不需要搜索源码，直接执行）：

1. **创建 Service Account**：
   - 登录：`POST /api/v1/auth`，body: `{"username":"admin","password":"xxx"}`，返回 Set-Cookie（accessToken/refreshToken）
   - 创建 SA：`POST /api/v1/service-accounts`，body: `{"username":"terraform-sa","roleBindings":[{"role":"EnvironmentAdmin","scopes":[]}]}`，用 cookie 认证
   - 返回 accessKey + secretKey，后续所有 Terraform/Go client 调用用这对凭证

2. **查询基础设施**（用 cookie 认证，所有路径带 `/api/v1` 前缀）：
   - 可用区：`GET /api/v1/providers/zones`
   - VPC 列表：`GET /api/v1/providers/vpcs`
   - 子网列表：`GET /api/v1/providers/vpcs/{vpcId}/subnets?zone={zone}`
   - K8s 集群：`GET /api/v1/providers/k8s-clusters`
   - K8s 节点组：`GET /api/v1/providers/k8s-clusters/{clusterId}/node-groups`
   - IAM 角色：`GET /api/v1/providers/roles`
   - 安全组：`GET /api/v1/providers/security-groups`

3. **查询已有插件**：
   - `GET /api/v1/connect-plugins?page=1&size=20&brief=true`（用 SA 签名认证）

4. **创建 Kafka 实例**（PoC 前置依赖）：
   - 通过 Terraform 创建，需要：environment_id, zone, subnet(private), version(5.3.8), deploy_type(IAAS), reserved_aku(6)
   - 实例创建后还需要：Kafka User + Topic + ACLs（TOPIC ALL, GROUP ALL, CLUSTER ALL, TRANSACTIONAL_ID ALL）

5. **K8s 集群访问权限**（需要用户操作）：
   - 创建 Connector 前，先通过 `GET /api/v1/providers/k8s-clusters/{clusterId}` 检查 `accessible` 字段
   - 如果 `accessible: false`，**提醒用户在 EKS 中添加 CMP 的访问条目**（这一步我无法自己完成）
   - 用户确认 accessible 变为 true 后再继续创建 Connector
   - **还需要确认 deploy profile 已配置**：`GET /api/v1/deploy-profiles` 不能返回空列表。如果为空，说明 CMP 的 providers 配置缺失，需要用户在 CMP 配置中添加（这是系统级配置，不是 API 操作）
   - 验证方式：`GET /api/v1/providers/k8s-clusters/{clusterId}?profile={profileName}` 应该返回集群详情而不是 "provider not found"

#### CMP REST API 参数格式（与 Terraform 字段名不同）

这些是直接调 CMP REST API 时的正确字段名（不是 Terraform 的字段名）：

**创建 Instance**: `POST /api/v1/instances`
```json
{"name":"xxx","version":"5.3.8","spec":{"deployType":"IAAS","reservedAku":6,"networks":[{"zone":"ap-southeast-1a","subnet":"subnet-xxx"}]},"features":{"walMode":"EBSWAL","security":{"authenticationMethods":["sasl"],"transitEncryptionModes":["plaintext"],"dataEncryptionMode":"NONE"}}}
```

**创建 User**: `POST /api/v1/instances/{id}/users`
```json
{"name":"connect-user","password":"xxx"}
```
注意：字段是 `name` 不是 `username`

**创建 Topic**: `POST /api/v1/instances/{id}/topics`
```json
{"name":"topic-name","partition":3,"compactStrategy":"DELETE","configs":[{"key":"retention.ms","value":"86400000"}]}
```
注意：`compactStrategy` 是必填（DELETE 或 COMPACT），`configs` 是 `[{key,value}]` 数组不是 map

**创建 ACL**: `POST /api/v1/instances/{id}/acls`
```json
{"params":[{"accessControlParam":{"user":"connect-user","operationGroup":"ALL","permissionType":"ALLOW"},"resourcePatternParam":{"resourceType":"TOPIC","name":"*","patternType":"LITERAL"}}]}
```
注意：ACL 是嵌套结构（accessControlParam + resourcePatternParam），user 字段不带 "User:" 前缀

**Produce Message**: `POST /api/v1/instances/{id}/topics/{topicId}/message-channels`
```json
{"messages":[{"key":"xxx","content":"xxx"}]}
```
注意：topicId 必须是 Base64 内部 ID，不是 topic name

- 所有 API 路径前缀：`/api/v1`
- Controller `@RequestMapping("/connectors")` → 实际路径 `/api/v1/connectors`
- 基础设施查询在 ProviderController：`/api/v1/providers/zones`、`/api/v1/providers/vpcs` 等
- Deploy Profiles：`/api/v1/deploy-profiles`（从 cmpConfig 读取，空环境可能返回空列表）
- 认证方式：Service Account 用签名认证（Terraform client），Console 用 cookie 认证（POST /api/v1/auth）

#### Benchmark 注意事项

- Go HTTP client 默认 30s 超时，benchmark 需要增加到 120s（`c.HTTPClient.Timeout = 120 * time.Second`）
- 必须在代码中设置 `os.Setenv("NO_PROXY", "*")` 防止代理干扰
- CMP Produce API 可能返回 `Cluster.Timeout`（Admin client timeout），这是 Kafka 实例负载过高的信号
- 如果遇到 Cluster.Timeout，**不要等待恢复**，直接删除所有 connector 和旧 Kafka instance，创建新的 instance。这比等待恢复更快更可靠
- **Cluster.Timeout 的根因是 AWS Spot 实例被回收**，不是 Kafka 过载。Spot 实例随时可能被回收，导致 broker 不可用。遇到 Cluster.Timeout 就直接重建 instance，不要排查
- Benchmark 之间需要留足够的冷却时间，避免 Kafka 实例过载
- 大量 connector 创建/删除会给 K8s 和 Kafka 带来压力，每个组合之间建议等待 2-3 分钟

#### Connector 创建超时检查

**关键原则：Connector 创建有超时限制。如果 Connector 状态在 8 分钟内仍然是 CREATING，必须主动排查问题，不能无限等待。**

排查步骤：
1. **检查 K8s Pod 状态**：`kubectl get pods -n default | grep connect-deployment-{connector-id}`
   - 如果 Pod 不存在或 Pending，检查 K8s 资源是否充足
   - 如果 Pod 是 Running 但 0/1 Ready，检查 init container 日志

2. **检查 Connect REST API 状态**：
   ```bash
   kubectl exec {connect-pod} -- curl -s http://localhost:8083/connectors/{connector-name}/status
   ```
   - 如果 connector.state 是 RUNNING 但 tasks[0].state 是 FAILED，查看 tasks[0].trace 错误信息

3. **常见 Task FAILED 原因及解决方案**：
   - `KafkaSchemaHistory 配置错误`：缺少 `schema.history.internal.kafka.bootstrap.servers`，需要添加完整的 schema.history.internal.* 配置
   - `SASL 认证失败`：端口错误（9092 是 PLAINTEXT，9102 是 SASL_PLAINTEXT），或 JAAS 配置格式错误
   - `数据库连接失败`：检查 hostname/port/user/password，确保网络可达
   - `权限不足`：检查数据库用户权限（如 MySQL 需要 REPLICATION SLAVE）

4. **如果需要重建 Connector**：
   - 删除失败的 Connector：`DELETE /api/v1/connectors/{id}`
   - 修复配置后重新创建
   - 不要在同一个失败的 Connector 上反复重试

### 4. 指标截图与可视化

- CMP Connect 详情页指标截图需要我自主完成，用户不会介入
- 方案：创建 connector 时配置 metric_exporter（Prometheus Remote Write），指标写入本地或远程 Prometheus
- 可以用 Prometheus + Grafana Docker 容器本地运行，通过 Grafana HTTP API 渲染图表并导出 PNG
- 或者用 promtool/prometheus 的 API 直接查询数据，用脚本生成图表
- 具体方案在首次执行时确定并记录到本文件

### 5. Git 分支命名规范

- 插件沉淀分支：`feat/plugin-{plugin-name}`（如 `feat/plugin-s3-sink`、`feat/plugin-debezium-mysql`）
- Commit message 格式：`feat(plugin): add {plugin-name} configuration template and docs`
- PR 标题：`feat(plugin): {Plugin Display Name} configuration template, PoC, and performance baseline`

### 6. 外部依赖

不同插件需要不同的外部资源做 PoC 和性能测试。

**我能自主解决的**（通过 Docker 本地启动）：

| 插件 | Docker 方案 |
|------|------------|
| Debezium MySQL | `docker run mysql:8.0`，启动时开启 binlog（`--server-id=1 --log-bin=mysql-bin --binlog-format=ROW`） |
| Debezium PostgreSQL | `docker run postgres:16`，启动后配置 logical replication（`wal_level=logical`） |
| JDBC Source/Sink | 同上，用 MySQL 或 PostgreSQL |
| MongoDB | `docker run mongo:7.0` |
| ClickHouse | `docker run clickhouse/clickhouse-server` |

**需要用户提供的**（SaaS 服务或 AWS 资源）：

| 插件 | 需要用户提供 |
|------|-------------|
| S3 Sink | S3 bucket 名称 + region + IAM Role ARN（或用测试环境已有的） |
| Snowflake | Snowflake 账号、warehouse、database、schema、用户名密码 |
| 其他 SaaS 类 | 对应服务的连接信息 |

**执行策略**：当我开始沉淀一个插件时，先判断外部依赖类型。能用 Docker 解决的直接启动；需要用户提供的，在开始时主动询问。

### 7. Confluent 文档配置名称映射

Confluent Managed Connector 的参数名可能和开源版不同（如 Confluent 用 `output.data.format`，开源版没有）。

**处理策略**：
- 以开源版插件源码中的 `ConfigDef` 为权威来源（不是 Confluent 文档）
- Confluent 文档作为参考，用于了解推荐配置和最佳实践
- PoC 测试时用开源版参数名
- 如果发现 Confluent 专有参数，记录在 Migration Guide 中，说明"此参数在开源版中不存在/名称不同"

### 8. 性能指标体系

Source 和 Sink Connector 关注的指标不同，不同插件还有特有指标。**Sink Connector 通用指标**：

| 指标 | 说明 | 来源 |
|------|------|------|
| sink lag（records） | 消费延迟，最核心 | Kafka Connect metrics / CMP 面板 |
| throughput（records/s, bytes/s） | 写入吞吐 | Kafka Connect metrics |
| task error rate | 任务错误率 | Kafka Connect metrics |
| CPU / 内存利用率 | Worker 资源使用 | K8s metrics / CMP 面板 |
| flush 延迟 | 下游写入延迟（如 S3 put latency） | 插件特有 metrics |

**Source Connector 通用指标**：

| 指标 | 说明 | 来源 |
|------|------|------|
| source records polled/s | 从源端拉取速率 | Kafka Connect metrics |
| source records written/s | 写入 Kafka 速率 | Kafka Connect metrics |
| poll latency | 单次 poll 延迟 | Kafka Connect metrics |
| task error rate | 任务错误率 | Kafka Connect metrics |
| CPU / 内存利用率 | Worker 资源使用 | K8s metrics / CMP 面板 |
| 端到端延迟 | 源端变更到 Kafka 的延迟 | 需要自行计算 |

**插件特有指标**（在步骤 4 源码分析时从 `MetricGroup` 中提取）：
- S3 Sink：S3 put latency、S3 put failure rate、file rotation count
- Debezium MySQL/PostgreSQL：binlog/WAL position、snapshot progress、schema change count
- MongoDB：document count、bulk write latency
- ClickHouse：batch insert latency、batch size

---

## SOP 步骤

### 步骤 1：选型决策

- 基于 Confluent 使用排名确定下一个插件
- 检查 `system-plugins.json` 确认该插件是否已注册
- 检查配置模板是否已存在（读取后端源码确认）
- 产出：更新插件优先级列表，确认目标插件

### 步骤 2：开源版本适配

- 在 Confluent Hub / GitHub 找到该插件的开源版本
- 确认与 AutoMQ Kafka Connect 版本（基于 Kafka 3.9.0）兼容
- 如果 CDN 上没有该插件包，告知用户需要上传
- 如果 `system-plugins.json` 中没有注册，准备 PR 添加
- 产出：兼容性验证报告 + 插件包 URL

### 步骤 3：PoC 验证

- **配置来源**：先从 Confluent 文档拉取该插件的推荐配置模板，用 Confluent 的配置作为 PoC 的基础
- 启动外部依赖（Docker 本地启动或询问用户获取 SaaS 连接信息）
- 在 terraform-provider-automq 中用 Terraform API 做端到端 PoC
- 创建 connector → **往 topic 写入测试数据** → **验证数据到达目标（如 S3 bucket 中有文件）**
- **PoC 验证标准：数据必须端到端流动，不是 task RUNNING 就算通过**
- 同时验证插件兼容性和 API 可用性
- 如果发现 Confluent 参数名与开源版不一致，记录映射关系
- 产出：PoC 测试报告（HCL 配置 + 测试结果 + 参数映射表 + 数据验证截图）

### 步骤 4：源码分析与配置理解

- **git clone 插件源码到本地**（从 GitHub）
- 找到继承 `Connector` 的主类，分析 `config()` 方法返回的 `ConfigDef`
- 从 `ConfigDef` 提取所有参数：名称、类型、默认值、文档、是否必填、是否敏感、分组
- 从 `MetricGroup` / metrics 注册代码提取插件特有指标
- AI 辅助分析：
  - 核心配置参数含义和影响
  - 配置与性能的关系（如 flush.size 对吞吐和延迟的影响）
  - 常见配置错误和最佳实践
  - 敏感参数识别
- 产出：配置参数分析文档（含 ConfigDef 完整参数表 + 特有指标列表）

### 步骤 5：性能基准测试

- 根据插件类型（Source/Sink）选择对应的指标体系（见上方"性能指标体系"）
- 在不同配置组合下做性能测试：
  - 配置 vs 吞吐/延迟关系曲线
  - 不同 Worker 规格（Tier）下的性能基准
  - 推荐配置（针对不同数据量级）
- 采集插件特有指标（步骤 4 中识别的）
- 可能需要在当前项目中增加性能测试模块/脚本
- 产出：性能报告（原始数据）

### 步骤 5a：可视化素材产出

- 性能数据表格（配置组合 × 吞吐/延迟/CPU/内存）
- 配置 vs 性能关系曲线图
- 不同 Tier 性能对比图
- CMP Connect 详情页指标面板截图（sink lag、throughput、task 状态等真实运行截图）
- 产出：图表 + 截图（嵌入文档，复用于官网/GTM/客户 PoC）

### 步骤 6：配置模板生成

- 基于以上分析生成配置模板：
  - 必填参数 + 推荐默认值
  - 可选参数 + 说明
  - 敏感参数标记
  - 参数校验规则
  - 参数分组（用于前端表单渲染）
- 录入后端配置模板系统（依赖配置模板 API ready）
- 产出：配置模板（JSON）

### 步骤 7：文档产出

每个插件配套文档集：

- Quick Start — 快速开始指南
- Configuration Reference — 配置参考（所有参数说明）
- Performance Tuning — 性能调优指南（含步骤 5 的数据和图表）
- Migration Guide — 从 MSK/Confluent 迁移指南
- Troubleshooting — 常见问题排查

这些文档作为官网文档链接，也作为前端表单中的 documentation_link。

### 步骤 8：PR 提交

- 创建分支 `feat/plugin-{plugin-name}`
- 提交所有产出：配置模板、文档、测试代码、system-plugins.json 更新
- 推送并创建 PR
- Commit message：`feat(plugin): add {plugin-name} configuration template and docs`

---

## 可复用的脚本和工具

以下脚本在 S3 Sink 插件沉淀过程中创建并验证过，后续插件沉淀可以复用（修改常量即可）：

### 环境管理

| 脚本 | 用途 | 关键点 |
|------|------|--------|
| `test-connector/reset_env.go` | 删除所有 connector 和 instance，创建新 instance | 遇到 Cluster.Timeout 时用这个重置环境 |
| `test-connector/create_fresh_instance.go` | 单独创建新 Kafka instance | API 字段：`spec.networks[].subnet`（单数，不是 subnets 数组） |
| `test-connector/check_connector.go` | 快速查看所有 connector 状态 | 一行输出：id name state |
| `test-connector/check_instance.go` | 查看 instance 详情 | |

### PoC 验证

| 脚本 | 用途 | 关键点 |
|------|------|--------|
| `test-connector/s3_sink_poc_verify.go` | 端到端 PoC 验证 | 可作为模板，改 connector 配置即可复用 |
| `test-connector/produce_and_verify.go` | 简单的 produce + 状态检查 | |

### 源码分析

| 脚本 | 用途 | 关键点 |
|------|------|--------|
| `test-connector/s3_sink_configdef_extract.go` | 从 Java ConfigDef 提取参数表 | 通用的 ConfigDef 解析器，改源码路径即可复用其他插件 |

### 性能测试

| 脚本 | 用途 | 关键点 |
|------|------|--------|
| `test-connector/setup_and_benchmark.go` | 完整的 setup + benchmark（创建 user/topic/ACL + 6 组合测试） | 最完整的版本，包含正确的 CMP API 参数格式 |
| `test-connector/s3_sink_full_benchmark.go` | 纯 benchmark runner（假设 instance/user/topic/ACL 已存在） | 需要手动更新 instanceID 和 topicID |

### 图表生成

用 Python matplotlib 生成图表（已验证可用）：
```bash
pip3 install --break-system-packages matplotlib  # macOS 需要 --break-system-packages
```
图表生成代码模板见 S3 Sink benchmark 的 matplotlib 脚本（在对话历史中，需要提取为独立脚本）。

### 关键常量模板

每次沉淀新插件时需要更新的常量：
```go
const (
    cmpEndpoint = "http://xxx:8080"        // 用户提供
    accessKey   = "xxx"                     // 用户提供或自己创建
    secretKey   = "xxx"                     // 用户提供或自己创建
    envID       = "env-xxx"                 // 用户提供
    instanceID  = "kf-xxx"                  // 创建后获取
    topicID     = "xxx"                     // 创建 topic 后从 GET /topics 获取（Base64 内部 ID）
    pluginID    = "conn-plugin-xxx"         // 从 GET /connect-plugins 获取或上传后获取
    k8sCluster  = "eks-xxx"                 // 从 GET /providers/k8s-clusters 获取
    iamRole     = "arn:aws:iam::xxx"        // 用户提供
    kafkaUser   = "connect-user"            // 固定
    kafkaPass   = "ConnectPoC123!"          // 固定
)
```
