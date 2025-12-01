---
page_title: "Provider: AutoMQ"
description: |-
    AutoMQ Provider is utilized to manage resources in AutoMQ environment. The provider allows for the management of instances and Kafka resources within those instances (such as Topics, Users, and ACLs).
---

# AutoMQ Provider
![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)

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
      source = "automq/automq"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint
  automq_byoc_access_key_id = var.automq_byoc_access_key_id
  automq_byoc_secret_key    = var.automq_byoc_secret_key
}

locals {
  enable_tls = contains(var.security_transit_encryption_modes, "tls")
  is_k8s     = var.deploy_type == "K8S"
}

resource "automq_kafka_instance" "k8s" {
  environment_id = var.environment_id
  name           = var.instance_name
  description    = var.description
  version        = var.automq_version

  compute_specs = {
    reserved_aku = var.reserved_aku
    deploy_type  = var.deploy_type

    networks                   = var.networks
    dns_zone                   = var.dns_zone
    data_buckets               = length(var.data_buckets) > 0 ? var.data_buckets : null
    instance_role              = var.instance_role
    kubernetes_cluster_id      = local.is_k8s ? var.kubernetes_cluster_id : null
    kubernetes_node_groups     = local.is_k8s ? var.kubernetes_node_groups : null
    kubernetes_namespace       = local.is_k8s ? var.kubernetes_namespace : null
    kubernetes_service_account = local.is_k8s ? var.kubernetes_service_account : null
  }

  features = {
    wal_mode = var.wal_mode

    instance_configs = var.instance_configs

    metrics_exporter = var.metrics_exporter

    table_topic = var.table_topic

    security = {
      authentication_methods   = var.security_authentication_methods
      transit_encryption_modes = var.security_transit_encryption_modes
      data_encryption_mode     = var.security_data_encryption_mode

      certificate_authority = local.enable_tls ? tls_self_signed_cert.example[0].cert_pem : null
      certificate_chain     = local.enable_tls ? tls_self_signed_cert.example[0].cert_pem : null
      private_key           = local.enable_tls ? tls_private_key.example[0].private_key_pem : null
    }
  }
}


resource "tls_private_key" "example" {
  count     = local.enable_tls ? 1 : 0
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "example" {
  count           = local.enable_tls ? 1 : 0
  private_key_pem = tls_private_key.example[0].private_key_pem

  subject {
    common_name  = "automq.private"
    organization = "AutoMQ"
  }

  validity_period_hours = 8760

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

output "instance_status" {
  description = "Provisioning status of the AutoMQ Kafka instance"
  value       = automq_kafka_instance.k8s.status
}

output "instance_id" {
  description = "Identifier of the AutoMQ Kafka instance"
  value       = automq_kafka_instance.k8s.id
}

output "instance_endpoints" {
  description = "Client connection endpoints"
  value       = automq_kafka_instance.k8s.endpoints
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `automq_byoc_access_key_id` (String, Sensitive) Set the Access Key Id of Service Account. You can create and manage Access Keys by using the AutoMQ Cloud BYOC Console. Learn more about AutoMQ Cloud BYOC Console access [here](https://docs.automq.com/automq-cloud/manage-identities-and-access/service-accounts).
- `automq_byoc_endpoint` (String) Set the AutoMQ BYOC environment endpoint. The endpoint looks like http://{hostname}:8080. You can get this endpoint when deploy environment complete.
- `automq_byoc_secret_key` (String, Sensitive) Set the Secret Access Key of Service Account. You can create and manage Access Keys by using the AutoMQ Cloud BYOC Console. Learn more about AutoMQ Cloud BYOC Console access [here](https://docs.automq.com/automq-cloud/manage-identities-and-access/service-accounts).

## Helpful Links/Information

* [Report Bugs](https://github.com/AutoMQ/terraform-provider-automq/issues)

* [Documents](https://www.automq.com)

* [Request Features](https://www.automq.com/contact)
