output "current_service_name" {
  value = module.automq_byoc.service_name
}

output "instance_ip" {
  value = module.automq_byoc.instance_ip
}

output "vpc_id" {
  value = module.automq_byoc.vpc_id
}

output "ebs_volume_id" {
  value = module.automq_byoc.ebs_volume_id
}

output "security_group_name" {
  value = module.automq_byoc.security_group_name
}

output "access_address" {
  value = module.automq_byoc.Service_access_address
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
  value = module.automq_byoc.cmp_role_arn
}

output "cmp_policy_arn" {
  value = module.automq_byoc.cmp_policy_arn
}

output "cmp_instance_profile_arn" {
  description = "Instance configuration file ARN:"
  value       = module.automq_byoc.cmp_instance_profile_arn
}

output "route53_zone_id" {
  value = module.automq_byoc.route53_zone_id
}