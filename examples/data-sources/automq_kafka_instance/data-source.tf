data "automq_kafka_instance" "example" {
  environment_id = local.env_id
  name           = "automq-example-1"
}

output "example-id" {
  value = data.automq_kafka_instance.example.id
}