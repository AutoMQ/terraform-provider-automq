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

resource "automq_kafka_acl" "example" {
  environment_id  = "example123"
  kafka_instance  = "kf-rrn5s50fzpr23urd"
  resource_type   = "TOPIC"
  resource_name   = "example-"
  pattern_type    = "PREFIXED"
  principal       = "automq_kafka_user"
  operation_group = "ALL"
  permission      = "ALLOW"
}
