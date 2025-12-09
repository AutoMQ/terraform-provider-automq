# Import format: <environment_id>@<instance_id>@<link_id>@<source_group_id>
# source_group_id - Consumer group identifier in the source cluster
terraform import automq_kafka_mirror_group.example env-abc123@kf-xyz789@link-sync@orders-group
