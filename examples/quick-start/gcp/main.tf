terraform {
  required_providers {
    automq = {
      source  = "automq/automq"
      version = "~> 0.4.4"
    }
  }
}

provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint
  automq_byoc_access_key_id = var.automq_byoc_access_key_id
  automq_byoc_secret_key    = var.automq_byoc_secret_key
}

resource "automq_kafka_instance" "gke" {
  environment_id = var.environment_id
  name           = var.instance_name
  description    = "AutoMQ on GKE Standard"
  version        = var.automq_version

  compute_specs = {
    reserved_aku = var.reserved_aku
    deploy_type  = "K8S"

    networks = [
      {
        zone    = var.zone
        subnets = []
      }
    ]

    kubernetes_cluster_id            = "projects/<project>/locations/<location>/clusters/<name>"
    instance_types                   = [var.instance_type]
    kubernetes_load_balancer_subnets = [var.load_balancer_subnet_id]
    schedule_spec                    = var.schedule_spec
  }

  features = {
    # EBSWAL is the API value for block-storage-backed WAL on GCP.
    wal_mode = "EBSWAL"

    security = {
      authentication_methods   = ["anonymous"]
      transit_encryption_modes = ["plaintext"]
    }
  }
}
