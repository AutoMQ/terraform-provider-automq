# outputs.tf
output "automq_byoc_env_id" {
  value = var.automq_byoc_env_id
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
  value = aws_iam_role.automq_byoc_role.arn
}

output "automq_byoc_policy_arn" {
  value = aws_iam_policy.automq_byoc_policy.arn
}

output "automq_byoc_vpc_route53_zone_id" {
  description = "The ID of the Route 53 zone"
  value       = aws_route53_zone.private.zone_id
}

output "automq_byoc_instance_profile_arn" {
  description = "The ARN of the instance profile for automq_byoc_service_role"
  value       = aws_iam_instance_profile.automq_byoc_instance_profile.arn
}

output "automq_byoc_env_webui_address" {
  value = "http://${aws_eip.web_ip.public_ip}:8080"
}

output "automq_byoc_env_console_ami" {
  value = var.specified_ami_by_marketplace ? data.aws_ami.marketplace_ami_details.id : var.automq_byoc_env_console_ami
}

output "automq_byoc_instance_id" {
  description = "AutoMQ BYOC Console instance ID."
  value       = aws_instance.web.id
}

output "automq_byoc_env_console_cidr" {
  description = "CIDR block to restrict the source IP address range for accessing the AutoMQ environment console."
  value       = var.automq_byoc_env_console_cidr
}