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