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

resource "automq_kafka_topic" "example" {
  environment_id = "example123"
  kafka_instance = "kf-rrn5s50fzpr23urd"
  name           = "example"
  partitions     = 16
  compact_strategy = "DELETE"
  config = {
    "retention.ms" = "86400000"
  }
}
