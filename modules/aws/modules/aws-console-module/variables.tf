variable "automq_byoc_env_id" {
  description = "This parameter is used to create resources within the environment. Additionally, all cloud resource names will incorporate this parameter as part of their names.This parameter supports only numbers, uppercase and lowercase English letters, and hyphens. It must start with a letter and is limited to a length of 32 characters."
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z][a-zA-Z0-9-]{0,31}$", var.automq_byoc_env_id)) && !can(regex("_", var.automq_byoc_env_id))
    error_message = "The environment_id must start with a letter, can only contain alphanumeric characters and hyphens, cannot contain underscores, and must be 32 characters or fewer."
  }
}

variable "cloud_provider_region" {
  description = "Set the cloud provider's region. AutoMQ will deploy to this region."
  type        = string
}

variable "automq_byoc_ec2_instance_type" {
  type    = string
  default = "c5.xlarge"
}

variable "automq_byoc_env_version" {
  description = "Set the version for the AutoMQ BYOC environment console. It is recommended to keep the default value, which is the latest version."
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

variable "automq_byoc_data_bucket_name" {
  description = "Set the existed object storage bucket for that used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket."
  type        = string
}

variable "automq_byoc_ops_bucket_name" {
  description = "Set the existed object storage bucket for that used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket."
  type        = string
}

variable "specified_ami_by_marketplace" {
  description = "Specifies the switch to obtain ami id from the cloud market. If true, it means to obtain the specified version of ami id through the cloud market. Otherwise, it needs to be specified manually."
  type = bool
  default = true
}

variable "automq_byoc_env_console_ami" {
  description = "When obtaining ami id from non-cloud market, manually specify ami id."
  type = string
}

variable "automq_byoc_env_console_cidr" {
  description = "Set CIDR block to restrict the source IP address range for accessing the AutoMQ environment console. If not set, the default is 0.0.0.0/0."
  type = string
  default = "0.0.0.0/0"
}