resource "automq_connect_cluster" "example" {
  environment_id = var.automq_environment_id
  name           = "connect-cluster-prod"
  description    = "Production Connect worker cluster"

  plugins = [
    {
      name    = "s3-sink"
      version = "11.1.0"
    }
  ]

  kafka_cluster = {
    kafka_instance_id = automq_kafka_instance.example.id
  }

  capacity = {
    type = "autoscaling"

    autoscaling = {
      worker_resource_spec = "TIER2"
      min_worker_count     = 2
      max_worker_count     = 6
      scale_in_policy = {
        cpu_utilization_percentage = 70
      }
      scale_out_policy = {
        cpu_utilization_percentage = 70
      }
    }
  }

  compute = {
    type = "k8s"

    kubernetes = {
      cluster_id      = var.kubernetes_cluster_id
      namespace       = "connect"
      service_account = "connect-sa"
    }

    iam_role = var.connect_iam_role
  }

  worker_config = {
    "offset.flush.interval.ms" = "10000"
  }

  tags = {
    team = "data-platform"
  }
}

variable "automq_environment_id" {
  type = string
}

variable "kubernetes_cluster_id" {
  type = string
}

variable "connect_iam_role" {
  type = string
}
