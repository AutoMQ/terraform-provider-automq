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
