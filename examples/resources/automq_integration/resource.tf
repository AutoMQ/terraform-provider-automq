resource "automq_integration" "example-1" {
  name = "example-1"
  type = "cloudWatch"
  cloudwatch_config = {
    namespace = "example"
  }
}

resource "automq_integration" "example-2" {
  name     = "example-2"
  type     = "kafka"
  endpoint = "http://xxxxx.xxx"
  kafka_config = {
    security_protocol = "SASL_PLAINTEXT"
    sasl_mechanism    = "PLAIN"
    sasl_username     = "example"
    sasl_password     = "example"
  }
}

resource "automq_integration" "example-3" {
  name     = "example-3"
  type     = "prometheus"
  endpoint = "http://xxxxx.xxx"
}
