terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
  }
}

locals {
  env_id = "example"

  automq_byoc_host          = "http://localhost:8081"
  automq_byoc_access_key_id = "goiNxB8DfbbXJ85B"
  automq_byoc_secret_key    = "QPyEIcBXHKOBzEeeCZcpNSMRjXtj4XiS"

  instance_deploy_region = "cn-hangzhou"
  instance_deploy_zone   = "cn-hangzhou-b"

  instance_deploy_subnet = "vsw-bp14v5eikr8wrgoqje7hr"
}

provider "automq" {
  automq_byoc_host          = local.automq_byoc_host
  automq_byoc_access_key_id = local.automq_byoc_access_key_id
  automq_byoc_secret_key    = local.automq_byoc_secret_key
}

resource "automq_integration" "example" {
  environment_id = local.env_id
  name           = "integration-example"
  type           = "cloudWatch"
  cloudwatch_config = {
    namespace = "example"
  }
}

resource "automq_kafka_instance" "example" {
  environment_id = local.env_id

  name           = "automq-example-1"
  description    = "example"
  cloud_provider = "aliyun"
  region         = local.instance_deploy_region
  networks = [
    {
      zone    = local.instance_deploy_zone
      subnets = [local.instance_deploy_subnet]
    }
  ]
  compute_specs = {
    aku     = "12"
    version = "1.1.0"
  }
  acl          = true
  integrations = [automq_integration.example.id]
}

resource "automq_kafka_topic" "example" {
  environment_id = local.env_id

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
  environment_id = local.env_id

  kafka_instance_id = automq_kafka_instance.example.id
  username          = "automq_kafka_user"
  password          = "automq_kafka_user"
}

resource "automq_kafka_acl" "example" {
  environment_id = local.env_id

  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TOPIC"
  resource_name   = automq_kafka_topic.example.name
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

