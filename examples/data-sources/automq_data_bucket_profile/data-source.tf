data "automq_deploy_profile" "test" {
  environment_id = var.automq_environment_id
  name           = "default"
}

data "automq_data_bucket_profiles" "test" {
  environment_id = var.automq_environment_id
  profile_name   = data.automq_deploy_profile.test.name
}

variable "automq_environment_id" {
  type = string
}
