# Import format: <environment_id>@<instance_id>@<link_id>@<source_topic_name>
# source_topic_name - Topic name in the source Kafka cluster being mirrored
terraform import automq_kafka_mirror_topic.example env-abc123@kf-xyz789@link-sync@orders
