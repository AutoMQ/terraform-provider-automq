resource "automq_kafka_topic" "example" {
  environment_id    = "env-example"
  kafka_instance_id = "kf-gm4q8xxxxxxvkg2"
  name              = "example"
  partition         = 16
  compact_strategy  = "DELETE"
  configs = {
    "delete.retention.ms" = "86400"
  }
}
