# S3 Sink Connector — Troubleshooting

## Scenario 1: Task FAILED Due to Converter Mismatch

### Symptoms

Connector creates successfully but task immediately enters FAILED state. Logs show:

```
ERROR WorkerSinkTask Task threw an uncaught and unrecoverable exception
Caused by: org.apache.kafka.connect.errors.DataException:
  Converting byte[] to Kafka Connect data failed due to serialization error
```

### Root Cause

The message key or value format does not match the Worker's converter configuration. Most commonly: the default `JsonConverter` cannot deserialize a plain string key.

### Resolution

1. Identify your message format (check how producers write to the topic)
2. Update `worker_config` to match:

```hcl
worker_config = {
  "key.converter"                  = "org.apache.kafka.connect.storage.StringConverter"
  "value.converter"                = "org.apache.kafka.connect.json.JsonConverter"
  "value.converter.schemas.enable" = "false"
}
```

3. Delete and recreate the connector with the corrected config

### Converter Selection Matrix

| Key Format | Value Format | key.converter | value.converter | schemas.enable |
| --- | --- | --- | --- | --- |
| String | JSON (no schema) | StringConverter | JsonConverter | false |
| JSON | JSON (no schema) | JsonConverter + schemas.enable=false | JsonConverter | false |
| String | JSON (with schema) | StringConverter | JsonConverter | true |
| String | Avro | StringConverter | AvroConverter | N/A |
| String | Plain text | StringConverter | StringConverter | N/A |
| Bytes | Bytes | ByteArrayConverter | ByteArrayConverter | N/A |

Full class names:

- `org.apache.kafka.connect.storage.StringConverter`
- `org.apache.kafka.connect.json.JsonConverter`
- `io.confluent.connect.avro.AvroConverter`
- `org.apache.kafka.connect.converters.ByteArrayConverter`

## Scenario 2: Data Not Appearing in S3

### Symptoms

Connector state is RUNNING with zero failed tasks, but no new files appear in the S3 bucket.

### Root Cause

Usually one of: no data in topic, flush.size not reached, missing time-based rotation, or IAM permission issues.

### Resolution

Check in this order:

1. **Verify topic has data** — Use CMP console or Produce API to confirm messages exist in the topic

2. **Check flush.size** — If `flush.size=5000` but only 100 messages exist, data won't flush until the count is reached per partition

3. **Check rotate.interval.ms** — If not configured, data only flushes when `flush.size` is reached. Add `rotate.interval.ms=600000` (10 min) to ensure time-based flushing

4. **Check IAM permissions** — The IAM Role needs these S3 permissions:

   ```json
   {
     "Effect": "Allow",
     "Action": [
       "s3:PutObject",
       "s3:GetBucketLocation",
       "s3:ListBucket",
       "s3:AbortMultipartUpload",
       "s3:ListMultipartUploadParts"
     ],
     "Resource": [
       "arn:aws:s3:::your-bucket",
       "arn:aws:s3:::your-bucket/*"
     ]
   }
   ```

5. **Verify bucket and region** — Confirm `s3.bucket.name` and `s3.region` match your actual S3 bucket

6. **Check connector logs** for S3 access errors:

   ```
   GET /api/v1/connectors/{connectorId}/logs?tailLines=100
   ```

## Scenario 3: KubernetesValidationFailed on Connector Creation

### Symptoms

```
Error: Create Connector Error
API Error: Connector.KubernetesValidationFailed
Message: Failed to validate Kubernetes prerequisites
```

### Root Cause

CMP cannot access the specified Kubernetes cluster, or the namespace/service account does not exist.

### Resolution

1. **Verify K8s cluster accessibility:**

   ```
   GET /api/v1/providers/k8s-clusters/{clusterId}
   ```

   Confirm `accessible: true` in the response.

2. **Check EKS access entries** — CMP needs IAM access to the EKS cluster. Ensure CMP's IAM Role is added to the EKS `aws-auth` ConfigMap or Access Entries.

3. **Check security groups** — EKS node security groups must allow inbound connections from CMP.

4. **Verify namespace exists** — The specified `kubernetes_namespace` must exist in the cluster.

5. **Verify service account exists** — The specified `kubernetes_service_account` must exist in the namespace.

## Scenario 4: High S3 Write Latency

### Symptoms

Data appears in S3 but with much higher delay than expected (e.g., minutes instead of seconds).

### Root Cause

Combination of large flush.size, missing time-based rotation, insufficient worker resources, or cross-region S3 writes.

### Resolution

1. **Reduce flush.size** — Smaller values trigger more frequent S3 writes:

   ```
   flush.size = 100    # for low-latency
   flush.size = 1000   # for balanced (default)
   ```

2. **Add rotate.interval.ms** — Ensures data is written even at low message rates:

   ```
   rotate.interval.ms = 60000   # 1 minute max delay
   ```

3. **Increase task_count** — More tasks = more parallel consumers:

   ```
   task_count = <number_of_partitions>
   ```

4. **Upgrade worker tier** — CPU-constrained workers process records slower:

   ```
   worker_resource_spec = "TIER2"   # or TIER3 for high volume
   ```

5. **Enable S3 Transfer Acceleration** — For cross-region writes:

   ```
   s3.wan.mode = true
   ```

6. **Check consumer lag** — High lag indicates the connector is falling behind. View lag on the CMP Connect detail page.

## Scenario 5: Connector FAILED with No Obvious Error

### Symptoms

Connector state changes to FAILED but the error message is generic or missing.

### Resolution

1. **Check connector logs via CMP API:**

   ```
   GET /api/v1/connectors/{connectorId}/logs?tailLines=200
   ```

   Look for `ERROR` or `WARN` entries, especially around task startup and S3 operations.

2. **Check worker resource usage** — If the worker pod is OOMKilled or CPU-throttled, tasks may fail silently. Upgrade to a higher tier:

   | Current Tier | Upgrade To | When |
   | --- | --- | --- |
   | TIER1 | TIER2 | OOMKilled or high CPU throttling |
   | TIER2 | TIER3 | Processing large messages or high volume |

3. **Verify Kafka ACLs are complete** — Missing ACLs can cause silent failures. Required ACLs:

   | Resource Type | Resource Name | Pattern Type | Operation |
   | --- | --- | --- | --- |
   | TOPIC | `*` or specific topic name | LITERAL | ALL |
   | GROUP | `connect` (prefix) | PREFIXED | ALL |
   | CLUSTER | `kafka-cluster` | LITERAL | ALL |
   | TRANSACTIONAL_ID | `connect` (prefix) | PREFIXED | ALL |

   Missing any of these can cause the connector to fail during startup or message consumption.

4. **Check Kafka connectivity** — Verify the Kafka instance is running and the SASL credentials are correct.

5. **Restart the connector** — Delete and recreate via Terraform if the issue is transient.

## Checking Connector Logs via CMP API

For any troubleshooting scenario, connector logs are your primary diagnostic tool:

```
GET /api/v1/connectors/{connectorId}/logs?tailLines=100
```

Response contains the most recent log lines from the connector worker pod. Look for:

- `ERROR` — Critical failures (converter errors, S3 access denied, OOM)
- `WARN` — Potential issues (retries, slow S3 responses)
- `DataException` — Converter mismatch (see Scenario 1)
- `AmazonS3Exception` — S3 permission or configuration errors
- `ConnectException` — Kafka connectivity or ACL issues

You can also check connector status programmatically:

```
GET /api/v1/connectors/{connectorId}
```

Key fields in the response:

- `state` — Should be `RUNNING`
- `failedTaskCount` — Should be `0`
- `runningTaskCount` — Should match your `task_count`

## Required Kafka ACLs (Complete List)

The S3 Sink Connector requires these Kafka ACLs to function correctly:

| Resource Type | Resource Name | Pattern Type | Operation | Purpose |
| --- | --- | --- | --- | --- |
| TOPIC | `*` or specific topic names | LITERAL | ALL | Read messages from source topics |
| GROUP | `connect` | PREFIXED | ALL | Consumer group for offset management |
| CLUSTER | `kafka-cluster` | LITERAL | ALL | Cluster-level operations (metadata, offsets) |
| TRANSACTIONAL_ID | `connect` | PREFIXED | ALL | Exactly-once delivery support |

Missing any ACL will cause connector failures. The GROUP and TRANSACTIONAL_ID resources use PREFIXED pattern type because Kafka Connect generates group and transaction IDs with the `connect` prefix.
