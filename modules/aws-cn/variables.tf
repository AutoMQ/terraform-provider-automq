variable "automq_byoc_env_name" {
  description = "automq_byoc service postfix"
  type        = string
  default     = "automq_byoc"

  # Added name restrictions, only uppercase and lowercase letters, numbers, and underscores can be used. The length is within 32 char
  validation {
    condition     = can(regex("^[a-zA-Z0-9-]{1,32}$", var.automq_byoc_env_name)) && !can(regex("_", var.automq_byoc_env_name))
    error_message = "The environment_id can only contain alphanumeric characters and hyphens, cannot contain underscores, and must be 32 characters or fewer."
  }
}

variable "create_vpc" {
  description = "Whether to create VPC"
  type        = bool
  default     = true
}

variable "automq_byoc_ec2_instance_type" {
  type    = string
  default = "c5.xlarge"
}

variable "aws_region" {
  description = "The AWS region to deploy in"
  type        = string
  default     = "cn-northwest-1"
}

variable "automq_byoc_vpc_id" {
  description = "The ID of the VPC"
  type        = string
}

variable "public_subnet_id" {
  description = "The ID of the subnet"
  type        = string
}

variable "automq_byoc_version" {
  description = "The version of automq_byoc"
  type        = string
}

variable "data_bucket_name" {
  description = "The name of the data bucket"
  type        = string
  default     = "automq-data"
}

variable "ops_bucket_name" {
  description = "The name of the ops bucket"
  type        = string
  default     = "automq-ops"
}

variable "create_data_bucket" {
  description = "Control whether to create the data bucket"
  type        = bool
  default     = true
}

variable "create_ops_bucket" {
  description = "Control whether to create the ops bucket"
  type        = bool
  default     = true
}

variable "specific_data_bucket_name" {
  description = "Specify the name of the automatically created data-bucket"
  type        = string
  default     = "automq-data"
}

variable "specific_ops_bucket_name" {
  description = "Specify the name of the automatically created ops-bucket"
  type        = string
  default     = "automq-ops"
}