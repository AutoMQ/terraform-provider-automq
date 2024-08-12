# outputs.tf
output "automq_byoc_env_name" {
  value = var.automq_byoc_env_name
}

output "automq_byoc_env_console_ec2_instance_ip" {
  value = aws_eip.web_ip.public_ip
}

output "automq_byoc_vpc_id" {
  value = var.automq_byoc_vpc_id
}

output "automq_byoc_env_console_public_subnet_id" {
  value = var.automq_byoc_env_console_public_subnet_id
}

output "ebs_volume_id" {
  value = [for bd in aws_instance.web.ebs_block_device : bd.volume_id][0]
}

output "automq_byoc_security_group_name" {
  value = aws_security_group.allow_all.name
}

output "automq_byoc_role_arn" {
  value = aws_iam_role.cmp_role.arn
}

output "automq_byoc_policy_arn" {
  value = aws_iam_policy.cmp_policy.arn
}

output "automq_byoc_vpc_route53_zone_id" {
  description = "The ID of the Route 53 zone"
  value       = aws_route53_zone.private.zone_id
}

output "automq_byoc_instance_profile_arn" {
  description = "The ARN of the instance profile for automq_byoc_service_role"
  value       = aws_iam_instance_profile.cmp_instance_profile.arn
}

output "AutoMQ_BYOC_Environment_WebUI_Address" {
  value = "Please wait for the service to initialize, about 1 min. Once ready, you can access the service at http://${aws_eip.web_ip.public_ip}:8080"
}

output "ami_id" {
  value = var.automq_byoc_env_version == "latest" ? data.aws_ami.latest_international_ami.id : data.aws_ami.specific_version_international_ami[0].id
}