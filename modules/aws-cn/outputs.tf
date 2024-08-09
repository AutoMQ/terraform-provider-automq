output "current_service_name" {
  value = module.cmp_service.service_name
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

output "access_address" {
  value = module.cmp_service.Service_access_address
}

output "data_bucket_name" {
  value = module.data_bucket.s3_bucket_id
}

output "data_bucket_arn" {
  value = module.data_bucket.s3_bucket_arn
}

output "ops_bucket_name" {
  value = module.ops_bucket.s3_bucket_id
}

output "ops_bucket_arn" {
  value = module.ops_bucket.s3_bucket_arn
}

output "cmp_role_arn" {
  value = module.cmp_service.cmp_role_arn
}

output "cmp_policy_arn" {
  value = module.cmp_service.cmp_policy_arn
}

output "cmp_instance_profile_arn" {
  description = "Instance configuration file ARN:"
  value       = module.cmp_service.cmp_instance_profile_arn
}

output "route53_zone_id" {
  value = module.cmp_service.route53_zone_id
}