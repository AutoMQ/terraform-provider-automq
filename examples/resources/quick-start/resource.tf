terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
  }
  required_version = ">= 0.1"
}

variable "env_id" {
  default = "env-example"
}

variable "user_password" {
  default   = "password"
  sensitive = true
}

provider "automq" {
  byoc_host       = "http://18.140.0.78:8080"
  byoc_access_key = "SvkmTn8lR0Rcwsd5"
  byoc_secret_key = "QJ3mbeImC7M4lVWqzptAHsgmzzGU76OD"
}

resource "automq_integration" "example" {
  environment_id = var.env_id
  endpoint       = "http://localhost:8082"
  name           = "integration-example"
  type           = "cloudWatch"
  cloudwatch_config = {
    namespace = "example"
  }
}

resource "automq_kafka_instance" "example" {
  environment_id = var.env_id

  name           = "automq-example-1"
  description    = "example"
  cloud_provider = "aws"
  region         = "ap-southeast-1"
  networks = [
    {
      zone    = "ap-southeast-1a"
      subnets = ["subnet-056e29f94d11b1414"]
    }
  ]
  compute_specs = {
    aku     = "6"
    version = "1.1.0"
  }
  acl = true
  integrations = [
    {
      integration_id   = automq_integration.example.id
      integration_type = automq_integration.example.type
    }
  ]
  // TODO 
}

resource "automq_kafka_topic" "example" {
  environment_id = var.env_id

  kafka_instance_id = automq_kafka_instance.example.id
  name              = "example"
  partition         = 64
  compact_strategy  = "DELETE"
  configs = {
    "delete.retention.ms" = "86400"
    "retention.ms"        = "3600000"
    "max.message.bytes"   = "1024"
  }
}

resource "automq_kafka_user" "example" {
  environment_id = var.env_id

  kafka_instance_id = automq_kafka_instance.example.id
  username          = "automq_kafka_user"
  password          = var.user_password
}

resource "automq_kafka_acl" "example" {
  environment_id = var.env_id

  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TOPIC"
  resource_name   = automq_kafka_topic.example.name
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

