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
  byoc_access_key = "VLaUIeNYndeOAXjaol32o4UAHvX8A7VE"
  byoc_secret_key = "CHlRi0hOIA8pAnzW"
  token           = "123456"
}

resource "automq_integration" "example" {
  environment_id = "example12"
  name           = "example11"
  type           = "cloudWatch"
  endpoint       = "http://localhost:8082"
  cloudwatch_config = {
    namespace = "example"
  }
}
