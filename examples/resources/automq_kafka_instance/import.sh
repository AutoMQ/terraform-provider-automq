# Import format: <environment_id>@<instance_id>
# environment_id - AutoMQ BYOC environment identifier (for example, env-abc123)
# instance_id    - Kafka instance ID returned by AutoMQ (for example, kf-xyz789)
terraform import automq_kafka_instance.example env-abc123@kf-xyz789
