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
  description = "实例配置文件 ARN: "
  value = module.cmp_service.cmp_instance_profile_arn
}