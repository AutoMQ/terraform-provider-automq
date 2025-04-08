resource "automq_integration" "example-1" {
  environment_id = "env-example"
  name           = "example-1"
  type           = "cloudWatch"
  cloudwatch_config = {
    namespace = "example"
  }
}

resource "automq_integration" "example-2" {
  environment_id = "env-example"
  name           = "example-2"
  type           = "prometheus"
  endpoint       = "http://xxxxx.xxx"
}

resource "automq_integration" "example-3" {
  environment_id = "env-example"
  name           = "example-3"
  type           = "prometheus_remote_write"
  endpoint       = "http://xxxxx.xxx"
  prometheus_remote_write_config = {
    auth_type = "basic"
    username = "username"
    password = "password"
  }
}
