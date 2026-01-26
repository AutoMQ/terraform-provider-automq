resource "automq_kafka_instance" "example" {
  environment_id = var.automq_environment_id
  name           = "automq-example"
  description    = "example deployment using inline compute specs"
  version        = "5.3.5"

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

    # security_groups = ["sg-example123"] # Optional. Omit this field entirely to let backend auto-generate. If specified, must contain at least one security group.
  }

  features = {
    wal_mode = "EBSWAL"
    security = {
      authentication_methods   = ["anonymous"]
      transit_encryption_modes = ["plaintext"]
    }
  }
}

variable "automq_environment_id" {
  type = string
}
