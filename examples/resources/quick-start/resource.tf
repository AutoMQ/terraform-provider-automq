terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
  }
}

locals {
  env_id = "cmp-dev"

  automq_byoc_endpoint      = "http://localhost:8081"
  automq_byoc_access_key_id = "9vBQQqxthaNJ9cCo"
  automq_byoc_secret_key    = "4REoWzOrU4l6dlZs1onK5Dbye6AxkGlJ"

  instance_deploy_region = "ap-southeast-1"
  instance_deploy_zone   = "ap-southeast-1a"

  instance_deploy_subnet = "subnet-06226ce8b221db030"
}

provider "automq" {
  automq_environment_id     = local.env_id
  automq_byoc_endpoint      = local.automq_byoc_endpoint
  automq_byoc_access_key_id = local.automq_byoc_access_key_id
  automq_byoc_secret_key    = local.automq_byoc_secret_key
}

resource "automq_integration" "example" {
  name = "integration-example-1"
  type = "cloudWatch"
  cloudwatch_config = {
    namespace = "example"
  }
}

resource "automq_kafka_instance" "example" {
  name           = "automq-example-2"
  description    = "example"
  cloud_provider = "aws"
  region         = local.instance_deploy_region
  networks = [
    {
      zone    = local.instance_deploy_zone
      subnets = [local.instance_deploy_subnet]
    }
  ]
  compute_specs = {
    aku     = "18"
    version = "1.1.0"
  }
  acl          = true
  integrations = [automq_integration.example.id]
  configs = {
    "auto.create.topics.enable" = "false"
    "log.retention.ms"          = "3600000"
  }
}

resource "automq_kafka_topic" "example" {
  kafka_instance_id = automq_kafka_instance.example.id
  name              = "example"
  partition         = 16
  configs = {
    "delete.retention.ms" = "86400"
    "retention.ms"        = "3600000"
    "max.message.bytes"   = "1024"
  }
}

resource "automq_kafka_user" "example" {
  kafka_instance_id = automq_kafka_instance.example.id
  username          = "automq_kafka_user-1"
  password          = "automq_kafka_user"
}

resource "automq_kafka_acl" "example" {
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TOPIC"
  resource_name   = automq_kafka_topic.example.name
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

