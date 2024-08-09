variable "automq_byoc_env_name" {
  description = "automq_byoc service postfix"
  type        = string
  default     = "automq_byoc"
}

variable "aws_region" {
  description = "The AWS region to deploy in"
  type        = string
  default     = "cn-northwest-1"
}

variable "automq_byoc_ec2_instance_type" {
  type    = string
  default = "c5.xlarge"
}

variable "automq_byoc_version" {
  description = "The version of automq_byoc"
  type        = string
}

variable "automq_byoc_vpc_id" {
  description = "The ID of the VPC"
  type        = string
}

variable "public_subnet_id" {
  description = "The ID of the subnet"
  type        = string
}

variable "data_bucket_name" {
  description = "The name of the data bucket"
  type        = string
}

variable "ops_bucket_name" {
  description = "The name of the ops bucket"
  type        = string
}