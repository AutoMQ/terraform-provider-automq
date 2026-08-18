# Import format: <environment_id>@<kafka_instance_id>@<user>|<resource_type>|<permission>|<resource_name>|<pattern_type>|<operation_group>
# Use the Kafka username without the "User:" prefix and URL-encode identity values containing reserved characters.
# Only ACLs with the wildcard host (`*`) can be imported because this resource does not expose a host argument.
terraform import automq_kafka_acl.topic 'env-abc123@kf-xyz789@orders-writer|TOPIC|ALLOW|orders|LITERAL|PRODUCE'
