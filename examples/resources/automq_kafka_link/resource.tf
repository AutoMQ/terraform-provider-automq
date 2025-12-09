resource "automq_kafka_link" "example" {
  environment_id    = var.environment_id
  instance_id       = var.instance_id
  link_id           = "example-link"
  start_offset_time = "latest"

  source_cluster = {
    endpoint          = "source-broker.example.com:9092"
    security_protocol = "SASL_SSL"
    sasl_mechanism    = "PLAIN"
    user              = "source-user"
    password          = var.source_password

    truststore_certificates = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIB...demo-truststore...IDAQAB
    -----END CERTIFICATE-----
    EOT

    keystore_certificate_chain = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIB...demo-keystore-cert...IDAQAB
    -----END CERTIFICATE-----
    EOT

    keystore_key = <<-EOT
    -----BEGIN PRIVATE KEY-----
    MIIE...demo-private-key...QAB
    -----END PRIVATE KEY-----
    EOT

    disable_endpoint_identification = false
  }
}

variable "environment_id" {
  type = string
}

variable "instance_id" {
  type = string
}

variable "source_password" {
  type      = string
  sensitive = true
}
