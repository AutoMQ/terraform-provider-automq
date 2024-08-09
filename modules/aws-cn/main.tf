provider "aws" {
  region = var.aws_region
}

# Conditional creation of data bucket
module "data_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.2"

  # Switch whether to create a bucket. If it is true, it will be created. If it is false, it will use the name entered by the user. If the name is empty, it will default to automq-data.
  create_bucket = var.create_data_bucket
  bucket        = var.create_data_bucket ? (
    var.specific_data_bucket_name == "" ? "automq-data" : var.specific_data_bucket_name
  ) : (
    var.data_bucket_name == "" ? "automq-data" : var.data_bucket_name
  )
  force_destroy = true
}

# Conditional creation of ops bucket
module "ops_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.2"

  create_bucket = var.create_ops_bucket
  bucket        = var.create_ops_bucket ? (
    var.specific_ops_bucket_name == "" ? "automq-ops" : var.specific_ops_bucket_name
  ) : (
    var.ops_bucket_name == "" ? "automq-ops" : var.ops_bucket_name
  )
  force_destroy = true
}

module "automq_byoc" {
  source = "./modules/aws-cn-module"

  aws_region       = var.aws_region
  aws_vpc_id       = var.aws_vpc_id
  subnet_id        = var.subnet_id
  aws_ami_id       = var.aws_ami_id
  data_bucket_name = module.data_bucket.s3_bucket_id
  ops_bucket_name  = module.ops_bucket.s3_bucket_id
  service_name     = var.service_name
}