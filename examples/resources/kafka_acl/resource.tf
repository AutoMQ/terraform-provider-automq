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
  byoc_access_key = "VLaUIeNYndeOAXjaol32o4UAHvX8A7VE"
  byoc_secret_key = "CHlRi0hOIA8pAnzW"
  token     = "123456"
}

resource "automq_kafka_acl" "example" {
  environment_id = "example123"
  kafka_instance_id = "kf-gm4q8tk1wqlavkg2"

  resource_type   = "TOPIC"
  resource_name   = "example-"
  pattern_type    = "PREFIXED"
  principal       = "User:automq_kafka_user"
  operation_group = "ALL"
  permission      = "ALLOW"
}
