terraform {
  required_providers {
    automq = {
      source = "hashicorp.com/edu/automq"
    }
  }
  required_version = ">= 0.1"
}

provider "automq" {
  byoc_host       = "http://localhost:8081"
  token           = "123456"
  byoc_access_key = "VLaUIeNYndeOAXjaol32o4UAHvX8A7VE"
  byoc_secret_key = "CHlRi0hOIA8pAnzW"
}

resource "automq_kafka_instance" "example" {
  environment_id = "example"
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
