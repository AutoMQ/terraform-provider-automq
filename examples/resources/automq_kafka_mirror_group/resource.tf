resource "automq_kafka_mirror_group" "example" {
  environment_id  = var.environment_id
  instance_id     = var.instance_id
  link_id         = var.link_id
  source_group_id = "orders-consumer"
}

variable "environment_id" {
  type = string
}

variable "instance_id" {
  type = string
}

variable "link_id" {
  description = "Identifier of the automq_kafka_link resource"
  type        = string
}
