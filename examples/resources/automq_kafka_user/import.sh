# Import format: <environment_id>@<kafka_instance_id>@<username>
# Set the desired password in configuration because AutoMQ never returns credentials.
terraform import automq_kafka_user.example env-abc123@kf-xyz789@service-user
