variable "service_name" {
  description = "cmp service postfix"
  type        = string
  default     = "cmp_service"
}

variable "aws_region" {
  description = "The AWS region to deploy in"
  type        = string
  default     = "cn-northwest-1"
}

variable "aws_vpc_id" {
  description = "The ID of the VPC"
  type        = string
}

variable "subnet_id" {
  description = "The ID of the subnet"
  type        = string
}

variable "aws_ami_id" {
  description = "The ID of the AMI to use"
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