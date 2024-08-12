output "automq_byoc_env_name" {
  value = module.automq_byoc.automq_byoc_env_name
}

output "automq_byoc_env_console_ec2_instance_ip" {
  value = module.automq_byoc.automq_byoc_env_console_ec2_instance_ip
}

output "automq_byoc_vpc_id" {
  value = module.automq_byoc.automq_byoc_vpc_id
}

output "automq_byoc_env_console_public_subnet_id" {
  value = module.automq_byoc.automq_byoc_env_console_public_subnet_id
}

output "automq_byoc_security_group_name" {
  value = module.automq_byoc.automq_byoc_security_group_name
}

output "AutoMQ_BYOC_Environment_WebUI_Address" {
  value = module.automq_byoc.AutoMQ_BYOC_Environment_WebUI_Address
}

output "automq_byoc_data_bucket_name" {
  value = module.automq_byoc_data_bucket_name.s3_bucket_id
}

output "automq_byoc_data_bucket_arn" {
  value = module.automq_byoc_data_bucket_name.s3_bucket_arn
}

output "automq_byoc_ops_bucket_name" {
  value = module.automq_byoc_ops_bucket_name.s3_bucket_id
}

output "automq_byoc_ops_bucket_arn" {
  value = module.automq_byoc_ops_bucket_name.s3_bucket_arn
}

output "automq_byoc_role_arn" {
  value = module.automq_byoc.automq_byoc_role_arn
}

output "automq_byoc_policy_arn" {
  value = module.automq_byoc.automq_byoc_policy_arn
}

output "automq_byoc_cmp_instance_profile_arn" {
  description = "Instance configuration file ARN"
  value       = module.automq_byoc.automq_byoc_instance_profile_arn
}

output "automq_byoc_vpc_route53_zone_id" {
  value = module.automq_byoc.automq_byoc_vpc_route53_zone_id
}

output "automq_byoc_console_ami_id" {
  value = module.automq_byoc.ami_id
}