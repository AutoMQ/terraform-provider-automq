# AutoMQ Connector Terraform API v2 — Schema（评审后终版）

> Issue: https://github.com/AutoMQ/automqbox/issues/5491
>
> 两层资源模型：`automq_connect_cluster`（Worker 集群）+ `automq_connector`（Connector 实例）。
> 插件管理参考 Strimzi：cluster 级别声明插件列表（不区分 managed/custom），connector 只声明 connector_class。

---

## 资源总览

| 资源 | 类型 | 职责 |
|------|------|------|
| `automq_connect_cluster` | resource | Worker 集群生命周期（插件集、compute、capacity、Kafka 连接、Worker 配置） |
| `automq_connector` | resource | Connector 实例（业务配置），挂载到 cluster |
| `automq_connector_plugin` | resource | 插件包上传到 CMP 插件仓库（不可变） |
| `automq_connector_plugin` | data source | 查询插件仓库中的插件（含系统内置） |

---

## automq_connect_cluster

管理 Kafka Connect Worker 集群。负责插件集、基础设施（compute）、容量（capacity）、Kafka 连接（kafka_cluster）和 Worker 级配置。

```hcl
resource "automq_connect_cluster" "main" {
  environment_id = "env-xxxxx"                                     # Required, ForceNew
  name           = "connect-cluster-prod"                          # Required
  description    = "Production Connect cluster"                    # Optional

  # ---- 插件集（参考 Strimzi spec.build.plugins） ----
  # 不区分 managed/custom，统一声明。
  # 变更插件列表会触发 Worker 滚动重启。
  # 同一插件不允许出现不同版本（创建时校验）。
  plugins = [
    {
      name    = "s3-sink"                                          # Required, 插件标识
      version = "11.1.0"                                           # Required
    },
    {
      name    = "debezium-mysql"
      version = "3.1.2"
    }
  ]

  # ---- Kafka 连接 ----
  # Worker 进程连接 Kafka broker 使用 CMP 内部鉴权，用户只需指定 Kafka 实例 ID。
  # 鉴权信息（security_protocol）已移至 automq_connector 资源，
  # 用于控制 connector 插件内部创建的 producer/consumer client 参数。
  kafka_cluster = {
    kafka_instance_id = "kf-xxxxx"                                 # Required, ForceNew
  }

  # ---- 容量（二选一：provisioned 或 autoscaling） ----
  capacity = {
    type = "provisioned"                                           # Required ("provisioned" | "autoscaling")

    # type = "provisioned" 时必填
    provisioned = {
      worker_resource_spec = "TIER2"                               # Required (TIER1/TIER2/TIER3/TIER4)
      worker_count         = 2                                     # Required, ≥1
    }

    # type = "autoscaling" 时必填
    # autoscaling = {
    #   worker_resource_spec = "TIER2"                             # Required
    #   min_worker_count     = 1                                   # Required, ≥1
    #   max_worker_count     = 4                                   # Required, ≥ min_worker_count
    #   scale_in_policy = {
    #     cpu_utilization_percentage = 20                           # Required
    #   }
    #   scale_out_policy = {
    #     cpu_utilization_percentage = 80                           # Required
    #   }
    # }
  }

  # ---- 计算资源 ----
  compute = {
    type = "k8s"                                                   # Required, ForceNew ("k8s" | "asg")

    # type = "k8s" 时必填, ForceNew
    kubernetes = {
      cluster_id      = "eks-xxxxx"                                # Required
      namespace       = "connect"                                  # Required
      service_account = "connect-sa"                               # Required
      scheduling_spec = "..."                                      # Optional (YAML snippet)
    }

    # type = "asg" 时必填, ForceNew（未来）
    # asg = {
    #   subnet_ids      = ["subnet-aaa", "subnet-bbb"]
    #   security_groups = ["sg-xxxxx"]
    # }

    iam_role = "arn:aws:iam::123456789012:role/connect-role"       # Optional, ForceNew
  }

  # ---- Worker 级配置 ----
  worker_config = {                                                # Optional
    "offset.flush.interval.ms" = "10000"
  }

  # ---- 指标导出 ----
  metric_exporter = {                                              # Optional
    remote_write_enable = true                                     # Required
    remote_write = {                                               # remote_write_enable=true 时必填
      endpoint   = "https://prometheus.example.com/api/v1/write"
      auth_type  = "basic"                                         # none | basic | bearer | sigv4
      username   = "prom-user"                                     # Optional (basic 时必填)
      password   = "***"                                           # Optional, Sensitive (basic 时必填)
      token      = "***"                                           # Optional, Sensitive (bearer 时必填)
      region     = "us-east-1"                                     # Optional (sigv4 时必填)
      prometheus_arn = "arn:aws:aps:..."                            # Optional (sigv4 时必填)
      labels     = { "env" = "prod" }                              # Optional
    }
  }

  tags    = { "team" = "data-platform" }                           # Optional
  version = "5.3.8"                                                # Optional (默认最新)

  timeouts = { create = "30m", update = "30m", delete = "20m" }   # Optional

  # ---- 只读字段 (Computed) ----
  # id                    = "connect-cluster-xxxxx"
  # state                 = "RUNNING"
  # kafka_connect_version = "3.9.0"
  # created_at            = "2026-03-30T12:00:00Z"
  # updated_at            = "2026-03-30T12:00:00Z"
}
```

### 关键行为

- 删除集群前必须先删除其上所有 Connector，否则拒绝（Q14）。
- 插件列表变更触发 Worker 滚动重启，期间 Connector 状态不受影响（Q20）。
- 同一插件不允许不同版本共存，创建/更新时前置校验（Q2）。
- Worker 强制使用 `connect.protocol=compatible`（incremental cooperative rebalance）（Q18）。
- 默认不允许 Connector 通过 override 自定义连接的 Kafka 集群（Q9）。
- 内部 Topic partition 数固定：offset=16, status=16, config=1（Q3）。

---

## automq_connector

管理单个 Connector 实例。挂载到一个 `automq_connect_cluster` 上运行。

```hcl
resource "automq_connector" "s3_sink" {
  environment_id     = "env-xxxxx"                                 # Required, ForceNew
  connect_cluster_id = automq_connect_cluster.main.id              # Required, ForceNew (设计上可变，第一期不支持)
  name               = "order-events-s3-sink"                      # Required, 全局唯一
  description        = "S3 sink for order events"                  # Optional
  connector_class    = "io.confluent.connect.s3.S3SinkConnector"   # Required, ForceNew
  task_count         = 3                                           # Required, ≥1

  # ---- Kafka 鉴权（用于 connector 插件内部的 producer/consumer client） ----
  # Worker 进程自身连接 Kafka 使用 CMP 内部鉴权，不需要用户提供。
  # 这里的 security_protocol 控制的是 connector 插件创建的 producer/consumer 的连接参数。
  kafka_cluster = {                                                # Optional
    security_protocol = {
      protocol         = "SASL_SSL"                                # Required
      username         = "connect-user"                            # Optional (SASL 时必填)
      password         = "***"                                     # Optional, Sensitive (SASL 时必填)
      sasl_mechanism   = "SCRAM-SHA-512"                           # Optional (默认 SCRAM-SHA-512)
      truststore_certs = "-----BEGIN CERTIFICATE-----..."         # Optional (默认用实例 CA)
      client_cert      = "-----BEGIN CERTIFICATE-----..."         # Optional (SSL mTLS 时必填)
      private_key      = "-----BEGIN PRIVATE KEY-----..."         # Optional, Sensitive (SSL mTLS 时必填)
    }
  }

  connector_config = {                                             # Optional
    "topics"         = "order-events"
    "s3.bucket.name" = "my-data-lake"
    "s3.region"      = "us-east-1"
    "flush.size"     = "1000"
  }

  connector_config_sensitive = {                                   # Optional, Sensitive
    "database.password" = "***"
  }

  initial_offsets = [                                              # Optional, Create-only
    {
      partition = { "server" = "server_01" }
      offset    = { "file" = "mysql-bin.000598", "pos" = "2326" }
    }
  ]

  timeouts = { create = "10m", update = "10m", delete = "10m" }   # Optional

  # ---- 只读字段 (Computed) ----
  # id             = "conn-xxxxx"
  # state          = "RUNNING"
  # connector_type = "SINK"
  # plugin_id      = "conn-plugin-xxxxx"
  # created_at     = "2026-03-30T12:00:00Z"
  # updated_at     = "2026-03-30T12:00:00Z"
}
```

### 关键行为

- Connector 名称全局唯一，为未来跨 cluster 迁移预留（Q11）。
- `connect_cluster_id` 第一期 ForceNew，未来支持原地迁移（Q11）。
- Cluster 未就绪时创建 Connector，CMP 排队等待（Q10）。
- initial_offsets 通过 Connect Offsets API（Kafka 3.6+）实现（Q12）。
- 删除 Connector 时不清理 offset；删除 Cluster 时清理（Q13）。
- Connector 和 Cluster 状态独立，不传导（Q15）。

---

## automq_connector_plugin（resource）

上传插件包到 CMP 插件仓库。不区分 managed/custom，统一模型。

```hcl
resource "automq_connector_plugin" "debezium_mysql" {
  environment_id     = "env-xxxxx"                                 # Required, ForceNew
  name               = "debezium-mysql"                            # Required, ForceNew
  version            = "2.5.0"                                     # Required, ForceNew
  storage_url        = "s3://bucket/debezium-mysql-2.5.0.zip"     # Required, ForceNew
  types              = ["SOURCE"]                                  # Required, ForceNew
  connector_class    = "io.debezium.connector.mysql.MySqlConnector"# Required, ForceNew
  description        = "Debezium MySQL CDC"                        # Optional, ForceNew
  documentation_link = "https://debezium.io/docs"                  # Optional, ForceNew

  timeouts = { create = "10m", delete = "10m" }                   # Optional

  # ---- 只读字段 (Computed) ----
  # id              = "conn-plugin-xxxxx"
  # plugin_provider = "CUSTOM"
  # status          = "ACTIVE"
  # created_at      = "2026-03-30T12:00:00Z"
  # updated_at      = "2026-03-30T12:00:00Z"
}
```

---

## automq_connector_plugin（data source）

查询插件仓库中的插件（含系统内置）。

```hcl
data "automq_connector_plugin" "s3_sink" {
  environment_id  = "env-xxxxx"
  connector_class = "io.confluent.connect.s3.S3SinkConnector"      # 按 connector_class 查询
}

# 输出：
# data.automq_connector_plugin.s3_sink.id
# data.automq_connector_plugin.s3_sink.name
# data.automq_connector_plugin.s3_sink.latest_version
# data.automq_connector_plugin.s3_sink.versions          (所有可用版本列表)
```

---

## 端到端示例

### 单租模式（provisioned，1 cluster : 1 connector）

```hcl
resource "automq_connect_cluster" "dedicated" {
  environment_id = "env-xxxxx"
  name           = "s3-sink-cluster"

  plugins = [
    { name = "s3-sink", version = "11.1.0" }
  ]

  kafka_cluster = {
    kafka_instance_id = automq_kafka_instance.main.id
  }

  capacity = {
    type = "provisioned"
    provisioned = {
      worker_resource_spec = "TIER2"
      worker_count         = 2
    }
  }

  compute = {
    type = "k8s"
    kubernetes = {
      cluster_id      = "eks-xxxxx"
      namespace       = "connect"
      service_account = "connect-sa"
    }
    iam_role = "arn:aws:iam::123456789012:role/connect-role"
  }
}

resource "automq_connector" "s3_sink" {
  environment_id     = "env-xxxxx"
  connect_cluster_id = automq_connect_cluster.dedicated.id
  name               = "order-s3-sink"
  connector_class    = "io.confluent.connect.s3.S3SinkConnector"
  task_count         = 3

  kafka_cluster = {
    security_protocol = {
      protocol = "SASL_SSL"
      username = "connect-user"
      password = var.connect_password
    }
  }

  connector_config = {
    "topics"         = "order-events"
    "s3.bucket.name" = "my-data-lake"
    "s3.region"      = "us-east-1"
    "flush.size"     = "1000"
  }
}
```

### 多租模式（autoscaling，1 cluster : N connectors）

```hcl
# 上传 custom 插件到仓库
resource "automq_connector_plugin" "debezium_mysql" {
  environment_id  = "env-xxxxx"
  name            = "debezium-mysql"
  version         = "2.5.0"
  storage_url     = "s3://my-plugins/debezium-mysql-2.5.0.zip"
  types           = ["SOURCE"]
  connector_class = "io.debezium.connector.mysql.MySqlConnector"
}

# 共享 Worker 集群，声明需要的插件集
resource "automq_connect_cluster" "shared" {
  environment_id = "env-xxxxx"
  name           = "shared-connect"

  plugins = [
    { name = "s3-sink",        version = "11.1.0" },
    { name = "debezium-mysql", version = "2.5.0"  }
  ]

  kafka_cluster = {
    kafka_instance_id = automq_kafka_instance.main.id
  }

  capacity = {
    type = "autoscaling"
    autoscaling = {
      worker_resource_spec = "TIER3"
      min_worker_count     = 1
      max_worker_count     = 4
      scale_in_policy  = { cpu_utilization_percentage = 20 }
      scale_out_policy = { cpu_utilization_percentage = 80 }
    }
  }

  compute = {
    type = "k8s"
    kubernetes = {
      cluster_id      = "eks-xxxxx"
      namespace       = "connect"
      service_account = "connect-sa"
    }
    iam_role = var.iam_role
  }
}

# S3 Sink Connector
resource "automq_connector" "s3_sink" {
  environment_id     = "env-xxxxx"
  connect_cluster_id = automq_connect_cluster.shared.id
  name               = "order-s3-sink"
  connector_class    = "io.confluent.connect.s3.S3SinkConnector"
  task_count         = 3

  kafka_cluster = {
    security_protocol = {
      protocol = "SASL_SSL"
      username = "connect-user"
      password = var.connect_password
    }
  }

  connector_config = {
    "topics"         = "order-events"
    "s3.bucket.name" = "my-data-lake"
    "s3.region"      = "us-east-1"
    "flush.size"     = "1000"
  }
}

# MySQL CDC Source Connector（同一个 cluster）
resource "automq_connector" "mysql_cdc" {
  environment_id     = "env-xxxxx"
  connect_cluster_id = automq_connect_cluster.shared.id
  name               = "mysql-cdc"
  connector_class    = "io.debezium.connector.mysql.MySqlConnector"
  task_count         = 1

  kafka_cluster = {
    security_protocol = {
      protocol = "SASL_SSL"
      username = "connect-user"
      password = var.connect_password
    }
  }

  connector_config = {
    "database.hostname"  = "db.example.com"
    "database.port"      = "3306"
    "database.user"      = "debezium"
    "database.server.id" = "1"
    "topic.prefix"       = "cdc"
  }

  connector_config_sensitive = {
    "database.password" = var.db_password
  }
}

### 单租到多租迁移（通过 initial_offsets）

```hcl
# 新的共享集群
resource "automq_connect_cluster" "new_shared" {
  # ...
}

# 迁移 Connector，通过 initial_offsets 指定上次消费位点
resource "automq_connector" "migrated_s3_sink" {
  environment_id     = "env-xxxxx"
  connect_cluster_id = automq_connect_cluster.new_shared.id
  name               = "order-s3-sink"
  connector_class    = "io.confluent.connect.s3.S3SinkConnector"
  task_count         = 3

  connector_config = {
    "topics"         = "order-events"
    "s3.bucket.name" = "my-data-lake"
    "s3.region"      = "us-east-1"
    "flush.size"     = "1000"
  }

  initial_offsets = [
    {
      partition = { "kafka_topic" = "order-events", "kafka_partition" = "0" }
      offset    = { "kafka_offset" = "1000000" }
    },
    {
      partition = { "kafka_topic" = "order-events", "kafka_partition" = "1" }
      offset    = { "kafka_offset" = "950000" }
    },
    {
      partition = { "kafka_topic" = "order-events", "kafka_partition" = "2" }
      offset    = { "kafka_offset" = "980000" }
    }
  ]
}
```
