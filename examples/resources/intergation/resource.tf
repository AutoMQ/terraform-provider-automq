terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
  }
  required_version = ">= 0.1"
}

provider "automq" {
  byoc_host = "http://localhost:8081"
  token     = "123456"
}

resource "automq_intergation" "example" {
  environment_id = "example123"
  name           = "example"
  type           = "Prometheus"
  endpoint       = "http://localhost:8081"
  prometheus_config = {
    "username" = "admin"
    "password" = "admin"
  }
}
