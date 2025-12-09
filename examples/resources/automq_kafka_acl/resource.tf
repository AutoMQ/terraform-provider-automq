variable "environment_id" {
  type = string
}

variable "kafka_instance_id" {
  type = string
}

resource "automq_kafka_acl" "topic" {
  environment_id    = var.environment_id
  kafka_instance_id = var.kafka_instance_id

  resource_type   = "TOPIC"
  resource_name   = "orders"
  pattern_type    = "LITERAL"
  principal       = "User:orders-writer"
  operation_group = "PRODUCE"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "group" {
  environment_id    = var.environment_id
  kafka_instance_id = var.kafka_instance_id

  resource_type   = "GROUP"
  resource_name   = "orders-consumers"
  pattern_type    = "PREFIXED"
  principal       = "User:orders-reader"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "cluster" {
  environment_id    = var.environment_id
  kafka_instance_id = var.kafka_instance_id

  resource_type   = "CLUSTER"
  resource_name   = "kafka-cluster"
  pattern_type    = "LITERAL"
  principal       = "User:platform-admin"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "transaction" {
  environment_id    = var.environment_id
  kafka_instance_id = var.kafka_instance_id

  resource_type   = "TRANSACTIONAL_ID"
  resource_name   = "payments-tx"
  pattern_type    = "LITERAL"
  principal       = "User:payments-service"
  operation_group = "ALL"
  permission      = "DENY"
}
