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

resource "automq_kafka_user" "example" {
  environment_id = "example123"
  kafka_instance_id = "kf-gm4q8tk1wqlavkg2"
  username       = "automq_kafka_user"
  password       = "123456789"
}
