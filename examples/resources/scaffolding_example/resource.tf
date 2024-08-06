terraform {
  required_providers {
    automq = {
      source  = "hashicorp.com/edu/automq"
    }
  }
  required_version = ">= 0.1"
}

provider "automq" {
  host = "http://localhost:8081"
  token = "123456"
}

resource "automq_kafka_instance" "example" {
  display_name = "example"
  description = "example"
  cloud_provider = "aliyun"
  region = "cn-hangzhou"
  network_type = "vpc"
  networks = [{
    zone = "cn-hangzhou-b"
    subnet = "vsw-bp14v5eikr8wrgoqje7hr"
  }]
  compute_specs = {
    aku = "6"
  }
}
