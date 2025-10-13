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


provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint
  automq_byoc_access_key_id = var.automq_byoc_access_key_id
  automq_byoc_secret_key    = var.automq_byoc_secret_key
}

resource "automq_kafka_instance" "example" {
  environment_id = var.automq_environment_id
  name           = "automq-example-vm"
  description    = "example"
  version        = "1.5.0"

  compute_specs = {
    reserved_aku = 3
    deploy_type  = "IAAS"
    provider     = "aws"
    region       = var.region
    vpc          = var.vpc_id
    networks = [
      {
        zone    = var.az
        subnets = [data.aws_subnets.aws_subnets_example.ids[0]]
      }
    ]
    data_buckets = [
      {
        bucket_name = "automq-data-bucket"
        provider    = "aws"
        region      = var.region
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
    metrics_exporter = {
      prometheus = {
        enabled   = true
        end_point = "http://prometheus.example.com/api/v1/write"
      }
    }
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
