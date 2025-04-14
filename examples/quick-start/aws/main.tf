terraform {
  required_providers {
    automq = {
      source = "automq/automq"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

data "aws_subnets" "aws_subnets_example" {
  provider = aws
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }
  filter {
    name   = "availability-zone"
    values = [var.az]
  }
}


resource "automq_integration" "prometheus_remote_write_example_1" {
  environment_id = var.automq_environment_id
  name           = "example-1"
  type           = "prometheusRemoteWrite"
  endpoint       = "http://example.com"
  deploy_profile = "default"

  prometheus_remote_write_config = {
    auth_type = "noauth"
  }
}

provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint
  automq_byoc_access_key_id = var.automq_byoc_access_key_id
  automq_byoc_secret_key    = var.automq_byoc_secret_key
}

data "automq_deploy_profile" "test" {
  environment_id = var.automq_environment_id
  name           = "default"
}

data "automq_data_bucket_profiles" "test" {
  environment_id = var.automq_environment_id
  profile_name   = data.automq_deploy_profile.test.name
}

resource "automq_kafka_instance" "example" {
  environment_id = var.automq_environment_id
  name           = "automq-example-vm"
  description    = "example"
  version        = "1.4.1"
  deploy_profile = data.automq_deploy_profile.test.name

  compute_specs = {
    reserved_aku = 3
    networks = [
      {
        zone    = var.az
        subnets = [data.aws_subnets.aws_subnets_example.ids[0]]
      }
    ]
    bucket_profiles = [
      {
        id = data.automq_data_bucket_profiles.test.data_buckets[0].id
      }
    ]
  }

  features = {
    wal_mode = "EBSWAL"
    security = {
      authentication_methods   = ["sasl"]
      transit_encryption_modes = ["plaintext"]
    }
    instance_configs = {
      "auto.create.topics.enable" = "false"
      "log.retention.ms"          = "3600000"
    }
    integrations = [
      automq_integration.prometheus_remote_write_example_1.id,
    ]
  }
}

resource "automq_kafka_topic" "example" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id
  name              = "topic-example"
  partition         = 16
  configs = {
    "delete.retention.ms" = "86400"
    "retention.ms"        = "3600000"
    "max.message.bytes"   = "1024"
  }
}

resource "automq_kafka_user" "example" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id
  username          = "kafka_user-example"
  password          = "user_password-example"
}

resource "automq_kafka_user" "example-1" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id
  username          = "kafka_user-example-1"
  password          = "user_password-example"
}


resource "automq_kafka_acl" "example-topic" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TOPIC"
  resource_name   = automq_kafka_topic.example.name
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "example-group" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "GROUP"
  resource_name   = "kafka_group-example"
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "example-cluster" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "CLUSTER"
  resource_name   = "kafka-cluster"
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example-1.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

resource "automq_kafka_acl" "example-transaction" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TRANSACTIONAL_ID"
  resource_name   = "kafka_transaction-example"
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example-1.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
}

variable "vpc_id" {
  type = string
}

variable "region" {
  type = string
}

variable "az" {
  type = string
}

variable "automq_byoc_endpoint" {
  type = string
}

variable "automq_byoc_access_key_id" {
  type = string
}

variable "automq_byoc_secret_key" {
  type = string
}

variable "automq_environment_id" {
  type = string
}