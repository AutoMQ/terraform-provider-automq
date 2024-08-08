provider "aws" {
  region = var.aws_region
}

module "data_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.2"

  create_bucket = var.create_data_bucket
  bucket        = var.data_bucket_name != "" ? var.data_bucket_name : "automq-data"
  force_destroy = true
}

module "ops_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.2"

  create_bucket = var.create_ops_bucket
  bucket        = var.ops_bucket_name != "" ? var.ops_bucket_name : "automq-ops"
  force_destroy = true
}

module "cmp_service" {
  source = "./modules/aws-cn-module"

  aws_region     = var.aws_region
  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
  aws_vpc_id     = var.aws_vpc_id
  subnet_id      = var.subnet_id
  aws_ami_id     = var.aws_ami_id
}