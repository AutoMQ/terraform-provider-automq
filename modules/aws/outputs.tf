output "automq_byoc_env_name" {
  description = "This parameter is used to create resources within the environment. Additionally, all cloud resource names will incorporate this parameter as part of their names.This parameter supports only numbers, uppercase and lowercase English letters, and hyphens. It must start with a letter and is limited to a length of 32 characters."
  value = module.automq_byoc.automq_byoc_env_name
}

output "automq_byoc_env_console_ec2_instance_ip" {
  description = "The instance IP of the deployed AutoMQ BYOC control panel. You can access the service through this IP."
  value = module.automq_byoc.automq_byoc_env_console_ec2_instance_ip
}

output "automq_byoc_vpc_id" {
  description = "AutoMQ BYOC is deployed in this VPC."
  value = module.automq_byoc.automq_byoc_vpc_id
}

output "automq_byoc_env_console_public_subnet_id" {
  description = "AutoMQ WebUI is deployed under this subnet."
  value = module.automq_byoc.automq_byoc_env_console_public_subnet_id
}

output "automq_byoc_security_group_name" {
  description = "Security group bound to the AutoMQ BYOC service."
  value = module.automq_byoc.automq_byoc_security_group_name
}

output "AutoMQ_BYOC_Environment_WebUI_Address" {
  description = "Address accessed by AutoMQ BYOC service"
  value = module.automq_byoc.AutoMQ_BYOC_Environment_WebUI_Address
}

output "automq_byoc_data_bucket_name" {
  description = "The object storage bucket for that used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket."
  value = module.automq_byoc_data_bucket_name.s3_bucket_id
}

output "automq_byoc_data_bucket_arn" {
  description = "Data storage bucket arn."
  value = module.automq_byoc_data_bucket_name.s3_bucket_arn
}

output "automq_byoc_ops_bucket_name" {
  description = "The object storage bucket for that used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket."
  value = module.automq_byoc_ops_bucket_name.s3_bucket_id
}

output "automq_byoc_ops_bucket_arn" {
  description = "Ops storage bucket arn."
  value = module.automq_byoc_ops_bucket_name.s3_bucket_arn
}

output "automq_byoc_role_arn" {
  description = "AutoMQ BYOC is bound to the role arn of the Console."
  value = module.automq_byoc.automq_byoc_role_arn
}

output "automq_byoc_policy_arn" {
  description = "AutoMQ BYOC is bound to a custom policy on the role arn."
  value = module.automq_byoc.automq_byoc_policy_arn
}

output "automq_byoc_cmp_instance_profile_arn" {
  description = "Instance configuration file ARN"
  value       = module.automq_byoc.automq_byoc_instance_profile_arn
}

output "automq_byoc_vpc_route53_zone_id" {
  description = "Route53 bound to the VPC."
  value = module.automq_byoc.automq_byoc_vpc_route53_zone_id
}

output "automq_byoc_console_ami_id" {
  description = "Mirror ami id of AutoMQ BYOC Console."
  value = module.automq_byoc.automq_byoc_ami_id
}

output "Automq_byoc_instance_id" {
  description = "AutoMQ BYOC Console instance ID."
  value = module.automq_byoc.automq_byoc_instance_id
}