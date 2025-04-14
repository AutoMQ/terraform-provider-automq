
resource "automq_integration" "prometheus_remote_write_example_1" {
  environment_id = var.automq_environment_id
  name           = "example-1"
  type           = "prometheusRemoteWrite"
  endpoint       = "http://example.com"
  deploy_profile = "default"

  prometheus_remote_write_config = {
    auth_type = "noauth"
  }
}

resource "automq_integration" "prometheus_remote_write_example_2" {
  environment_id = var.automq_environment_id
  name           = "example-2"
  type           = "prometheusRemoteWrite"
  endpoint       = "http://example.com"
  deploy_profile = "default"

  prometheus_remote_write_config = {
    auth_type = "basic"
    username  = "example-username"
    password  = "example-password"
  }
}

resource "automq_integration" "prometheus_remote_write_example_3" {
  environment_id = var.automq_environment_id
  name           = "example-3"
  type           = "prometheusRemoteWrite"
  endpoint       = "http://example.com"
  deploy_profile = "default"

  prometheus_remote_write_config = {
    auth_type    = "bearer"
    bearer_token = "example-token"
  }
}


variable "automq_environment_id" {
  type = string
}
