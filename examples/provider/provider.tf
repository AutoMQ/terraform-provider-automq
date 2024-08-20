terraform {
  required_providers {
    automq = {
      source = "automq/automq"
    }
  }
}

provider "automq" {
  automq_environment_id     = var.automq_environment_id      # By default, other resources will use the environment ID in this provider configuration

  automq_byoc_endpoint      = var.automq_byoc_endpoint       # optionally use AUTOMQ_BYOC_HOST environment variable
  automq_byoc_access_key_id = var.automq_byoc_access_key_id  # optionally use AUTOMQ_BYOC_ACCESS_KEY_ID environment variable
  automq_byoc_secret_key    = var.automq_byoc_secret_key     # optionally use AUTOMQ_BYOC_SECRET_KEY environment variable
}

variable "automq_byoc_endpoint" {
  type = string
}

variable "automq_byoc_access_key_id" {
  type = string
}

variable "automq_byoc_secret_key" {
  type = string
}

variable "automq_environment_id" {
  type = string
}