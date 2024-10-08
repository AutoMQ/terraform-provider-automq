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

locals {
  vpc_id = "vpc-0xxxxxxxxxxxf"
  region = "us-east-1"
  az     = "us-east-1b"
}

provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint
  automq_byoc_access_key_id = var.automq_byoc_access_key_id
  automq_byoc_secret_key    = var.automq_byoc_secret_key
}

data "aws_subnets" "aws_subnets_example" {
  provider = aws
  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }
  filter {
    name   = "availability-zone"
    values = [local.az]
  }
}

resource "automq_kafka_instance" "example" {
  environment_id = var.automq_environment_id
  name           = "automq-example-1"
  description    = "example"
  cloud_provider = "aws"
  region         = local.region
  networks = [
    {
      zone    = local.az
      subnets = [data.aws_subnets.aws_subnets_example.ids[0]]
    }
  ]
  compute_specs = {
    aku = "12"
  }
  acl = true
  configs = {
    "auto.create.topics.enable" = "false"
    "log.retention.ms"          = "3600000"
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

resource "automq_kafka_acl" "example" {
  environment_id    = var.automq_environment_id
  kafka_instance_id = automq_kafka_instance.example.id

  resource_type   = "TOPIC"
  resource_name   = automq_kafka_topic.example.name
  pattern_type    = "LITERAL"
  principal       = "User:${automq_kafka_user.example.username}"
  operation_group = "ALL"
  permission      = "ALLOW"
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