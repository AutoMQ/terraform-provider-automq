resource "automq_kafka_acl" "example-topic" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TOPIC"
  resource_name   = automq_kafka_topic.example.name
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "example-group" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "GROUP"
  resource_name   = "kafka_group-example"
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "example-cluster" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "CLUSTER"
  resource_name   = "kafka-cluster"
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "example-transaction" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TRANSACTIONAL_ID"
  resource_name   = "kafka_transaction-example"
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}
