data "automq_kafka_instance" "example" {
  environment_id = var.automq_environment_id
  name           = "automq-example-1"
}

variable "automq_environment_id" {
  type = string
}

output "example-id" {
  value = data.automq_kafka_instance.example.id
}