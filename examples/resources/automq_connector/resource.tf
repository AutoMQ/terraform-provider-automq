resource "automq_connector" "example" {
  environment_id             = var.automq_environment_id
  name                       = "my-s3-sink"
  description                = "S3 sink connector for order events"
  plugin_id                  = "plugin-s3-sink"
  kubernetes_cluster_id      = "k8s-cluster-id"
  kubernetes_namespace       = "kafka-connect"
  kubernetes_service_account = "connect-sa"
  task_count                 = 2
  version                    = "1.2.0"

  capacity {
    worker_count         = 2
    worker_resource_spec = "TIER2"
  }

  kafka_cluster {
    kafka_instance_id = automq_kafka_instance.example.id
    security_protocol {
      security_protocol = "SASL_PLAINTEXT"
      username          = "connect-user"
      password          = var.connect_password
    }
  }

  connector_config = {
    "topics"             = "orders"
    "s3.bucket.name"     = "my-data-lake"
    "s3.region"          = "us-east-1"
    "flush.size"         = "1000"
    "rotate.interval.ms" = "60000"
    "storage.class"      = "io.confluent.connect.s3.storage.S3Storage"
    "format.class"       = "io.confluent.connect.s3.format.json.JsonFormat"
    "partitioner.class"  = "io.confluent.connect.storage.partitioner.DefaultPartitioner"
  }

  labels = {
    team = "data-platform"
  }
}

variable "automq_environment_id" {
  type = string
}

variable "connect_password" {
  type      = string
  sensitive = true
}
