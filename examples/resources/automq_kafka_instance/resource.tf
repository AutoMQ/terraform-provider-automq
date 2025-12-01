resource "automq_kafka_instance" "example" {
  environment_id = var.automq_environment_id
  name           = "automq-example"
  description    = "example deployment using inline compute specs"
  version        = "1.5.0"

  compute_specs = {
    reserved_aku = 6
    deploy_type  = "IAAS"

    networks = [
      {
        zone    = "us-east-1a"
        subnets = ["subnet-aaaaaa"]
      }
    ]

    data_buckets = [
      {
        bucket_name = "automq-data-bucket"
      }
    ]
  }

  features = {
    wal_mode = "EBSWAL"
    security = {
      authentication_methods   = ["sasl"]
      transit_encryption_modes = ["tls"]
    }

    metrics_exporter = {
      prometheus = {
        auth_type = "noauth"
        endpoint  = "http://prometheus.example.com/api/v1/write"
        labels = {
          "env" = "test"
        }
      }
    }

    table_topic = {
      warehouse     = "default"
      catalog_type  = "HIVE"
      metastore_uri = "thrift://hive-metastore.example.com:9083"
    }
  }
}

variable "automq_environment_id" {
  type = string
}
