---
page_title: "Provider: AutoMQ"
description: |-
    AutoMQ Provider is utilized to manage resources in AutoMQ environment. The provider allows for the management of instances and Kafka resources within those instances (such as Topics, Users, and ACLs).
---

# AutoMQ Provider
![General_Availability](https://img.shields.io/badge/Lifecycle_Stage-General_Availability(GA)-green?style=flat&logoColor=8A3BE2&labelColor=rgba)

## Prerequisites

The AutoMQ environment represents a namespace, with each environment containing a complete set of AutoMQ control plane and data plane. All control and data planes of the AutoMQ BYOC environment are deployed within the user's VPC to ensure data privacy and security.

The AutoMQ Provider is used to manage an already installed AutoMQ environment. Therefore, before using the AutoMQ Provider, you need to complete the environment installation and obtain the access point and initial account information.

Refer to the following for specific operations:
- Install the AutoMQ environment by following the [documentation](https://registry.terraform.io/modules/AutoMQ/automq-byoc-environment/aws/latest).
- Create a ServiceAccount and obtain an AccessKey: After the environment is installed, users need to access the AutoMQ environment console through web browser, create a Service Account, and use the Access Key of the Service Account. Please refer to [document](https://docs.automq.com/automq-cloud/manage-identities-and-access/service-accounts).


## Example Usage
### Example Provider Configuration

```terraform
terraform {
  required_providers {
    automq = {
      source = "automq/automq"
    }
  }
}

provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint      # optionally use AUTOMQ_BYOC_ENDPOINT environment variable
  automq_byoc_access_key_id = var.automq_byoc_access_key_id # optionally use AUTOMQ_BYOC_ACCESS_KEY_ID environment variable
  automq_byoc_secret_key    = var.automq_byoc_secret_key    # optionally use AUTOMQ_BYOC_SECRET_KEY environment variable
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
```

### Example Usage for an AWS BYOC environment

```terraform
terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

locals {
  vpc_id = "vpc-03d6cb79151dbdfa3"
  region = "us-east-1"
  az     = "us-east-1a"
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
  name           = "automq-example"
  description    = "example"
  version        = "1.4.0"
  deploy_profile = data.automq_deploy_profile.test.name

  compute_specs = {
    reserved_aku = 3
    networks = [
      {
        zone    = local.az
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `automq_byoc_access_key_id` (String) Set the Access Key Id of Service Account. You can create and manage Access Keys by using the AutoMQ Cloud BYOC Console. Learn more about AutoMQ Cloud BYOC Console access [here](https://docs.automq.com/automq-cloud/manage-identities-and-access/service-accounts).
- `automq_byoc_endpoint` (String) Set the AutoMQ BYOC environment endpoint. The endpoint looks like http://{hostname}:8080. You can get this endpoint when deploy environment complete.
- `automq_byoc_secret_key` (String) Set the Secret Access Key of Service Account. You can create and manage Access Keys by using the AutoMQ Cloud BYOC Console. Learn more about AutoMQ Cloud BYOC Console access [here](https://docs.automq.com/automq-cloud/manage-identities-and-access/service-accounts).

## Helpful Links/Information

* [Report Bugs](https://github.com/AutoMQ/terraform-provider-automq/issues)

* [Documents](https://www.automq.com)

* [Request Features](https://www.automq.com/contact)
