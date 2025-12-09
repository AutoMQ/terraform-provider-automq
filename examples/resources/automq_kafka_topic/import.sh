# Import format: <environment_id>@<kafka_instance_id>@<topic_id>
# kafka_instance_id - Parent Kafka instance ID (for example, kf-xyz789)
# topic_id          - Topic identifier visible in the AutoMQ console or Terraform state
terraform import automq_kafka_topic.example env-abc123@kf-xyz789@topic-uuid
