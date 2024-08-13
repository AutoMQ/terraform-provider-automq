provider "aws" {
  region = var.cloud_provider_region
}

# Conditional creation of data bucket
module "automq_byoc_data_bucket_name" {
  source = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.2"

  # Switch whether to create a bucket. If it is true, it will be created. If it is false, it will use the name entered by the user. If the name is empty, it will default to automq-data.
  create_bucket = var.create_automq_byoc_data_bucket
  bucket        = var.create_automq_byoc_data_bucket ? (
    var.specific_data_bucket_name == "" ? "automq-data-${var.automq_byoc_env_id}" : var.specific_data_bucket_name
  ) : (
    var.automq_byoc_data_bucket_name == "" ? "automq-data-${var.automq_byoc_env_id}" : var.automq_byoc_data_bucket_name
  )
  force_destroy = true
}

# Conditional creation of ops bucket
module "automq_byoc_ops_bucket_name" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.2"

  create_bucket = var.create_automq_byoc_ops_bucket
  bucket        = var.create_automq_byoc_ops_bucket ? (
    var.specific_ops_bucket_name == "" ? "automq-ops-${var.automq_byoc_env_id}" : var.specific_ops_bucket_name
  ) : (
    var.automq_byoc_ops_bucket_name == "" ? "automq-ops-${var.automq_byoc_env_id}" : var.automq_byoc_ops_bucket_name
  )
  force_destroy = true
}

data "aws_availability_zones" "available" {}

module "automq_byoc_vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  count = var.create_new_vpc ? 1 : 0

  name = "automq-byoc-vpc-${var.automq_byoc_env_id}"
  cidr = "10.0.0.0/16"

  azs = slice(data.aws_availability_zones.available.names, 0, 3)
  public_subnets = ["10.0.0.0/20"]
  private_subnets = ["10.0.128.0/20", "10.0.144.0/20", "10.0.160.0/20"]

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Terraform   = "true"
    Environment = "dev"
  }
}

resource "aws_eip" "nat" {
  count = 3

  domain = "vpc"
}

resource "aws_security_group" "endpoint_sg" {
  count = var.create_new_vpc ? 1 : 0

  name        = "automq-byoc-endpoint-sg-${var.automq_byoc_env_id}"
  description = "Security group for VPC endpoint"
  vpc_id      = module.automq_byoc_vpc[0].vpc_id

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "automq-byoc-endpoint-sg-${var.automq_byoc_env_id}"
  }
}

resource "aws_vpc_endpoint" "ec2" {
  count = var.create_new_vpc ? 1 : 0

  vpc_id            = module.automq_byoc_vpc[0].vpc_id
  service_name      = "com.amazonaws.${var.cloud_provider_region}.ec2"
  vpc_endpoint_type = "Interface"
  security_group_ids = [aws_security_group.endpoint_sg[0].id]
  subnet_ids        = module.automq_byoc_vpc[0].private_subnets

  private_dns_enabled = true

  tags = {
    Name = "automq-byoc-ec2-endpoint-${var.automq_byoc_env_id}"
  }
}

resource "aws_vpc_endpoint" "s3" {
  count = var.create_new_vpc ? 1 : 0

  vpc_id            = module.automq_byoc_vpc[0].vpc_id
  service_name      = "com.amazonaws.${var.cloud_provider_region}.s3"
  vpc_endpoint_type = "Gateway"

  route_table_ids = concat(
    module.automq_byoc_vpc[0].public_route_table_ids,
    module.automq_byoc_vpc[0].private_route_table_ids
  )

  tags = {
    Name = "automq-byoc-s3-endpoint-${var.automq_byoc_env_id}"
  }
}

# Determine the vpc and subnet id, mainly related to the set variables of whether to create a VPC
locals {
  automq_byoc_vpc_id                       = var.create_new_vpc ? module.automq_byoc_vpc[0].vpc_id : var.automq_byoc_vpc_id
  automq_byoc_env_console_public_subnet_id = var.create_new_vpc ? element(module.automq_byoc_vpc[0].public_subnets, 0) : var.automq_byoc_env_console_public_subnet_id
}

module "automq_byoc" {
  source = "./modules/aws-console-module"

  cloud_provider_region                    = var.cloud_provider_region
  automq_byoc_vpc_id                       = local.automq_byoc_vpc_id
  automq_byoc_env_console_public_subnet_id = local.automq_byoc_env_console_public_subnet_id
  automq_byoc_data_bucket_name             = module.automq_byoc_data_bucket_name.s3_bucket_id
  automq_byoc_ops_bucket_name              = module.automq_byoc_ops_bucket_name.s3_bucket_id
  automq_byoc_env_id                       = var.automq_byoc_env_id
  automq_byoc_ec2_instance_type            = var.automq_byoc_ec2_instance_type
  automq_byoc_env_version                  = var.automq_byoc_env_version
  specified_ami_by_marketplace             = var.specified_ami_by_marketplace
  automq_byoc_env_console_ami              = var.automq_byoc_env_console_ami
}