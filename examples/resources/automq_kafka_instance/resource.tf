data "automq_deploy_profile" "default" {
  environment_id = "env-example"
  name           = "default"
}

data "automq_data_bucket_profiles" "test" {
  environment_id = "env-example"
  profile_name   = data.automq_deploy_profile.test.name
}

resource "automq_kafka_instance" "test" {
  environment_id = "env-example"
  name           = "example-1"
  description    = "example"
  deploy_profile = data.automq_deploy_profile.default.name
  version        = "1.4.0"

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
        id = data.automq_data_bucket_profiles.test.data_buckets[0].id
      }
    ]
  }

  features = {
    wal_mode = "EBSWAL"
    security = {
      authentication_methods   = ["anonymous"]
      transit_encryption_modes = ["plaintext"]
    }
  }
}

resource "automq_kafka_instance" "test" {
  environment_id = "env-example"
  name           = "example-1"
  description    = "example"
  deploy_profile = data.automq_deploy_profile.default.name
  version        = "1.4.0"

  compute_specs = {
    reserved_aku = 6
    kubernetes_node_groups = [{
      id = "k8s-node-group-1"
    }]
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
      transit_encryption_modes = ["tls"]
      data_encryption_mode     = "CPMK"
      certificate_authority    = file("${path.module}/certificate.pem")
      certificate_chain        = file("${path.module}/certificate.pem")
      private_key              = file("${path.module}/private_key.pem")
    }
  }
}

