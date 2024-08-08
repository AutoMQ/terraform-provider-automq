provider "aws" {
  region = var.aws_region
}

module "cmp_service" {
  source = "./modules/aws-cn-module"

  aws_region     = var.aws_region
  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
  aws_vpc_id     = var.aws_vpc_id
  aws_ami_id     = var.aws_ami_id
}

output "instance_ip" {
  value = module.cmp_service.instance_ip
}

output "vpc_id" {
  value = module.cmp_service.vpc_id
}

output "ebs_volume_id" {
  value = module.cmp_service.ebs_volume_id
}

output "security_group_name" {
  value = module.cmp_service.security_group_name
}

output "access_message" {
  value = module.cmp_service.access_message
}