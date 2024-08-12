variable "automq_byoc_env_name" {
  description = "This parameter is used to create resources within the environment. Additionally, all cloud resource names will incorporate this parameter as part of their names.This parameter supports only numbers, uppercase and lowercase English letters, and hyphens. It must start with a letter and is limited to a length of 32 characters."
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z][a-zA-Z0-9-]{0,31}$", var.automq_byoc_env_name)) && !can(regex("_", var.automq_byoc_env_name))
    error_message = "The environment_id must start with a letter, can only contain alphanumeric characters and hyphens, cannot contain underscores, and must be 32 characters or fewer."
  }
}

variable "create_new_vpc" {
  description = "This setting determines whether to create a new VPC. If set to true, a new VPC spanning three availability zones will be automatically created, which is recommended only for POC scenarios. For production scenario using AutoMQ, you should provide the VPC where the current Kafka application resides and check the current VPC against the requirements specified in the documentation."
  type        = bool
  default     = true
}

variable "automq_byoc_ec2_instance_type" {
  description = "Can be specified, But you need to ensure that 2 cores are 8g or above"
  type    = string
  default = "c5.xlarge"
}

variable "cloud_provider_region" {
  description = "Set the cloud provider's region. AutoMQ will deploy to this region."
  type        = string
}

variable "automq_byoc_vpc_id" {
  description = "The ID of the VPC"
  type        = string
}

variable "automq_byoc_env_console_public_subnet_id" {
  description = "Select a subnet for deploying the AutoMQ BYOC environment console. Ensure that the chosen subnet supports public access."
  type        = string
}

variable "automq_byoc_env_version" {
  description = "Set the version for the AutoMQ BYOC environment console. It is recommended to keep the default value, which is the latest version."
  type        = string
}

variable "automq_byoc_data_bucket_name" {
  description = "Set the existed object storage bucket for that used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket."
  type        = string
}

variable "automq_byoc_ops_bucket_name" {
  description = "Set the existed object storage bucket for that used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket."
  type        = string
}

variable "create_automq_byoc_data_bucket" {
  description = "This parameter controls whether to create a new bucket. If it is a POC scenario or there is no available bucket, set it to true. If there is already a suitable bucket, set it to false."
  type        = bool
  default     = true
}

variable "create_automq_byoc_ops_bucket" {
  description = "This parameter controls whether to create a new bucket. If it is a POC scenario or there is no available bucket, set it to true. If there is already a suitable bucket, set it to false."
  type        = bool
  default     = true
}

variable "specific_data_bucket_name" {
  description = "Set up a new object storage bucket that will be used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket."
  type        = string
  default     = "automq-data"
}

variable "specific_ops_bucket_name" {
  description = "Set up a new object storage bucket that will be used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket."
  type        = string
  default     = "automq-ops"
}