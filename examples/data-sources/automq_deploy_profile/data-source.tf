provider "automq" {
  automq_byoc_endpoint      = ""
  automq_byoc_access_key_id = ""
  automq_byoc_secret_key    = ""
}

data "automq_deploy_profile" "test" {
  environment_id = "env-example"
  name           = "default"
}
