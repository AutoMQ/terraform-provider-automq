terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
  }
}


locals {
  env_id = "example"

  automq_byoc_host          = "http://localhost:8081"
  automq_byoc_access_key_id = "RSaIMzrFC0kAmS1x"
  automq_byoc_secret_key    = "msnGqOuaV5gblXPvkWfxg7Ao7Nq2iyMo"
}

provider "automq" {
  automq_byoc_host          = local.automq_byoc_host
  automq_byoc_access_key_id = local.automq_byoc_access_key_id
  automq_byoc_secret_key    = local.automq_byoc_secret_key
}

resource "automq_kafka_acl" "example" {
  environment_id    = local.env_id
  kafka_instance_id = "kf-gm4q8tk1wqlavkg2"

  resource_type   = "TOPIC"
  resource_name   = "example-"
  pattern_type    = "PREFIXED"
  principal       = "User:automq_kafka_user"
  operation_group = "ALL"
  permission      = "ALLOW"
}
