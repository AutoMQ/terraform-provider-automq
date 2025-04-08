
data "automq_deploy_profile" "default" {
  environment_id = "env-example"
  name = "default"
}

resource "automq_kafka_instance" "test" {
    environment_id = "env-example"
    name          = "example-1"
    description   = "example"
    deploy_profile = data.automq_deploy_profile.default.name
    version       = "1.4.0"

    compute_specs = {
        reserved_aku = 6
        networks = [
            {
                zone    = "us-east-1a"
                subnets = ["subnet-xxxxxx"]
            }
        ]
        bucket_profiles = [
            {
                id = data.automq_deploy_profile.default.data_buckets.0.id
            }
        ]
    }

    features = {
        wal_mode = "EBSWAL"
        security = {
          authentication_methods = ["anonymous"]
          transit_encryption_modes = ["plaintext"]
        }
    }
}

variable "instance_deploy_zone" {
  type = string
}

variable "instance_deploy_subnet" {
  type = string
}