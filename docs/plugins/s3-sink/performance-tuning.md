# S3 Sink Connector — Performance Tuning Guide

## Test Environment

| Parameter | Value |
|-----------|-------|
| AutoMQ Version | 5.3.8 |
| Kafka Connect Version | 3.9.0 |
| Plugin | automq-kafka-connect-s3-11.1.0 |
| Region | ap-southeast-1 |
| Instance | 6 AKU, IAAS deployment |
| Topic | 3 partitions |
| Benchmark Date | 2026-04-02 |

## Kafka Produce Throughput Baseline

Measured using `kafka-producer-perf-test` running inside the K8s cluster (native Kafka protocol, no HTTP overhead).

![Produce Throughput by Record Size](benchmark-produce-throughput.png)

| Record Size | Records/sec | MB/sec | Avg Latency | P99 Latency |
|------------|-------------|--------|-------------|-------------|
| 200 bytes | 65,876 | 12.56 | 45 ms | 101 ms |
| 500 bytes | 53,706 | 25.61 | 94 ms | 113 ms |
| 1 KB | 39,746 | 38.81 | 177 ms | 302 ms |

![Produce Latency by Record Size](benchmark-produce-latency.png)

**Key insight**: Smaller records achieve higher records/sec, but larger records achieve higher data throughput (MB/sec).

## Connector Consumption Performance

Measured by pre-loading 10,000 JSON messages and observing how fast the S3 Sink Connector drains the consumer lag.

| flush.size | Tier | Records Consumed in 60s | Approx Rate (rec/sec) | Observation |
|-----------|------|------------------------|----------------------|-------------|
| 100 | TIER1 | ~9,800 | ~163 | Near-complete consumption. Small batches = frequent S3 PUTs. |
| 1000 | TIER1 | ~8,000 | ~133 | Good balance. Fewer S3 PUTs than flush.size=100. |
| 5000 | TIER1 | ~5,000 | ~83 | Partial consumption — flush.size > records per partition, data held in buffer. |
| 100 | TIER2 | ~35,100 (from backlog) | ~585 | Higher CPU enables faster processing of accumulated backlog. |
| 1000 | TIER2 | ~8,000 | ~133 | Similar to TIER1 at this message rate. |
| 5000 | TIER2 | ~5,000 | ~83 | Same flush.size threshold effect as TIER1. |

### Key Finding: flush.size vs Partition Record Count

When `flush.size` exceeds the number of records per partition, the connector holds records in memory waiting for the threshold. With 10,000 messages across 3 partitions (~3,333 per partition) and `flush.size=5000`, the connector never reaches the flush threshold by record count — data is only flushed when `rotate.interval.ms` triggers.

**Recommendation**: Set `flush.size` ≤ expected records per partition between flush intervals.

## flush.size Tuning

| flush.size | Estimated File Size (200B msgs) | S3 PUT Frequency | Recommended Scenario |
|-----------|--------------------------------|-----------------|---------------------|
| 100 | ~20 KB | Very frequent | Low-latency, near-real-time analytics |
| 1000 | ~200 KB | Moderate | Balanced (recommended default) |
| 5000 | ~1 MB | Low | High-throughput batch processing |
| 10000 | ~2 MB | Very low | Large-scale archival, batch ETL |

- Production default: `flush.size=1000`
- Always pair with `rotate.interval.ms` to ensure data freshness at low message rates
- Smaller flush.size = more S3 PUTs = higher S3 API cost but lower latency
- Larger flush.size = fewer, bigger files = better for downstream batch processing (Athena, Spark)

## Worker Spec Selection

| Tier | CPU | Memory | Recommended Use Case |
|------|-----|--------|---------------------|
| TIER1 | 0.5 | 1 GiB | Dev/test, low-volume topics (< 1K msgs/sec) |
| TIER2 | 1 | 2 GiB | Production starting point (1K-5K msgs/sec, recommended) |
| TIER3 | 2 | 4 GiB | High-volume production (5K-20K msgs/sec) |
| TIER4 | 4 | 8 GiB | Ultra-high throughput (> 20K msgs/sec) |

Throughput ranges are estimates. Actual throughput depends on message size, format, compression, S3 latency, and task count.

## task_count and worker_count Sizing

- `task_count` should not exceed the topic's partition count
- Recommended: `task_count = partition_count` for maximum parallelism
- Recommended: `worker_count = ceil(task_count / 2)` as starting point

## Compression

| s3.compression.type | CPU Overhead | Storage Savings |
|--------------------|-------------|-----------------|
| `none` | None | 0% |
| `gzip` | Moderate | 60-80% (recommended for most workloads) |

## Recommended Configuration Profiles

### Low-Latency (Real-Time Analytics)

```hcl
flush.size = 100
rotate.interval.ms = 60000
s3.compression.type = none
Worker: TIER1, 1 worker
```

### Balanced (Recommended Default)

```hcl
flush.size = 1000
rotate.interval.ms = 600000
s3.compression.type = gzip
Worker: TIER2, 2 workers
```

### High-Throughput (Batch ETL)

```hcl
flush.size = 10000
rotate.interval.ms = 3600000
s3.compression.type = gzip
Worker: TIER3, 4 workers
```

## Further Reading

- [Benchmark Results](benchmark-results.md) — Full benchmark data
- [Configuration Reference](configuration-reference.md) — Complete parameter documentation
- [Troubleshooting](troubleshooting.md) — Common issues and solutions
