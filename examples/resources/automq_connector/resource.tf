resource "automq_connector" "example" {
  environment_id     = var.automq_environment_id
  connect_cluster_id = automq_connect_cluster.example.id
  name               = "orders-s3-sink"
  description        = "S3 sink connector for order events"
  connector_class    = "io.confluent.connect.s3.S3SinkConnector"
  task_count         = 2

  kafka_cluster = {
    security_protocol = {
      protocol       = "SASL_PLAINTEXT"
      username       = "connect-user"
      password       = var.connect_password
      sasl_mechanism = "SCRAM-SHA-512"
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

  connector_config_sensitive = {
    "aws.secret.access.key" = var.aws_secret_access_key
  }
}

variable "automq_environment_id" {
  type = string
}

variable "connect_password" {
  type      = string
  sensitive = true
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
