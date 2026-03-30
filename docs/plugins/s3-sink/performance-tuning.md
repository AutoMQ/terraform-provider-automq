# S3 Sink Connector — Performance Tuning Guide

## 测试环境

| 项目 | 值 |
|------|-----|
| AutoMQ 版本 | 5.3.5 |
| Kafka Connect 版本 | 3.9.0 |
| Worker 规格 | TIER1 (0.5 CPU / 1 GiB) |
| Worker 数量 | 1 |
| Task 数量 | 1 |
| 插件版本 | automq-kafka-connect-s3-11.1.0 |
| 输出格式 | JsonFormat |
| Region | ap-southeast-1 |

## Produce 吞吐基准

通过 CMP API produce 消息（含 HTTP 签名开销），消息大小约 150 bytes/条。

| 总消息数 | 批量大小 | 耗时 | 吞吐 (msgs/sec) | 错误数 |
|---------|---------|------|-----------------|-------|
| 100 | 1 | 9.2s | 10.9 | 0 |
| 100 | 10 | 1.1s | 87.0 | 0 |
| 100 | 50 | 0.5s | 198.2 | 0 |
| 500 | 1 | 43.7s | 11.4 | 0 |
| 500 | 10 | 5.6s | 89.1 | 0 |
| 500 | 50 | 2.2s | 226.2 | 0 |
| 1000 | 1 | 87.2s | 11.5 | 0 |
| 1000 | 10 | 11.3s | 88.8 | 0 |
| 1000 | 50 | 4.4s | 226.8 | 0 |

注意：以上吞吐是 CMP API 的 produce 速率，包含 HTTP 签名和网络开销。Kafka 原生 producer 的吞吐会高得多。

## Connector 处理能力

- 1600 条消息（累计 100+500+1000）全部成功处理，零错误
- Connector 状态始终保持 RUNNING
- flush.size=3 配置下，每 3 条记录触发一次 S3 PUT

## 关键配置与性能关系

### flush.size

`flush.size` 控制每个 partition 累积多少条记录后 flush 到 S3。

| flush.size | 文件大小（估算） | S3 PUT 频率 | 适用场景 |
|-----------|----------------|------------|---------|
| 3 | ~450 bytes | 非常频繁 | 测试/调试 |
| 100 | ~15 KB | 频繁 | 低延迟场景 |
| 1000 | ~150 KB | 中等 | 平衡场景（推荐） |
| 5000 | ~750 KB | 较低 | 高吞吐场景 |
| 10000 | ~1.5 MB | 低 | 批量处理场景 |

建议：
- 生产环境推荐 `flush.size=1000`，平衡文件大小和延迟
- 如果对延迟敏感，配合 `rotate.interval.ms=60000`（1 分钟）确保数据及时写入
- 如果追求高吞吐，可以增大到 5000-10000

### rotate.interval.ms

时间触发 flush，即使 `flush.size` 未达到也会写入 S3。

| rotate.interval.ms | 说明 | 适用场景 |
|-------------------|------|---------|
| 60000 (1 min) | 最多 1 分钟延迟 | 实时分析 |
| 600000 (10 min) | 最多 10 分钟延迟 | 准实时（推荐） |
| 3600000 (1 hour) | 最多 1 小时延迟 | 批量 ETL |

### Worker 规格选择

| 规格 | CPU | 内存 | 适用场景 |
|------|-----|------|---------|
| TIER1 | 0.5 | 1 GiB | 低吞吐（< 1000 msgs/sec），测试环境 |
| TIER2 | 1 | 2 GiB | 中等吞吐（1000-5000 msgs/sec），推荐起步 |
| TIER3 | 2 | 4 GiB | 高吞吐（5000-20000 msgs/sec） |
| TIER4 | 4 | 8 GiB | 超高吞吐（> 20000 msgs/sec） |

### task_count 与 partition 的关系

- `task_count` 不应超过 topic 的 partition 数量
- 每个 task 消费一个或多个 partition
- 增加 task_count 可以提高并行度，但需要配合增加 worker_count

推荐：`task_count = partition_count`，`worker_count = ceil(task_count / 2)`

### 压缩

| s3.compression.type | CPU 开销 | 存储节省 | 适用场景 |
|--------------------|---------|---------|---------|
| none | 无 | 0% | 低 CPU 场景 |
| gzip | 中等 | 60-80% | 存储成本敏感（推荐） |

## 推荐配置

### 低延迟场景（实时分析）

```
flush.size = 100
rotate.interval.ms = 60000
s3.compression.type = none
```

### 平衡场景（推荐）

```
flush.size = 1000
rotate.interval.ms = 600000
s3.compression.type = gzip
```

### 高吞吐场景（批量 ETL）

```
flush.size = 10000
rotate.interval.ms = 3600000
s3.compression.type = gzip
s3.part.size = 52428800
```
