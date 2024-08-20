resource "automq_kafka_acl" "example" {
  environment_id    = "env-example"
  kafka_instance_id = "kf-gm4xxxxxxxxg2"

  resource_type   = "TOPIC"
  resource_name   = "example-"
  pattern_type    = "PREFIXED"
  principal       = "User:automq_xxxx_user"
  operation_group = "ALL"
  permission      = "ALLOW"
}
