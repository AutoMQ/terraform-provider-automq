resource "automq_kafka_mirror_topic" "example" {
  environment_id    = var.environment_id
  instance_id       = var.instance_id
  link_id           = var.link_id
  source_topic_name = "orders"
  state             = "PAUSED"
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
