# Import format: <environment_id>@<instance_id>@<link_id>
# instance_id - Kafka instance that owns the link (for example, kf-xyz789)
# link_id     - Identifier provided when the link was created
terraform import automq_kafka_link.example env-abc123@kf-xyz789@example-link
