variable "aws_region" {
  description = "The AWS region to deploy in"
  type        = string
  default     = "cn-northwest-1"
}

variable "aws_access_key" {
  description = "The AWS access key"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "The AWS secret key"
  type        = string
  sensitive   = true
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
}

variable "ops_bucket_name" {
  description = "The name of the ops bucket"
  type        = string
}