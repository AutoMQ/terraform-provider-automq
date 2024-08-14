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

resource "automq_kafka_instance" "example" {
  environment_id = local.env_id
  name           = "example123"
  description    = "example"
  cloud_provider = "aliyun"
  region         = "cn-hangzhou"
  networks = [{
    zone   = "cn-hangzhou-b"
    subnet = "vsw-bp14v5eikr8wrgoqje7hr"
  }]
  compute_specs = {
    aku     = "12"
    version = "1.1.0"
  }
}
