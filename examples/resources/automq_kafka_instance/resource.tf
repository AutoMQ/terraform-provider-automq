resource "automq_kafka_instance" "example" {
  environment_id = "env-example"
  name           = "automq-example-1"
  description    = "example"
  cloud_provider = "aws"
  region         = local.instance_deploy_region
  networks = [
    {
      zone    = var.instance_deploy_zone
      subnets = [var.instance_deploy_subnet]
    }
  ]
  compute_specs = {
    aku = "18"
  }
  acl = true
  configs = {
    "auto.create.topics.enable" = "false"
    "log.retention.ms"          = "3600000"
  }
}

variable "instance_deploy_zone" {
  type = string
}

variable "instance_deploy_subnet" {
  type = string
}