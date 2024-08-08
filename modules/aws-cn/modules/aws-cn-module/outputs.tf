# outputs.tf

output "instance_ip" {
  value = aws_eip.web_ip.public_ip
}

output "vpc_id" {
  value = data.aws_vpc.selected.id
}

output "ebs_volume_id" {
  value = [for bd in aws_instance.web.ebs_block_device : bd.volume_id][0]
}

output "security_group_name" {
  value = aws_security_group.allow_all.name
}

output "cmp_role_arn" {
  value = aws_iam_role.cmp_role.arn
}

output "cmp_policy_arn" {
  value = aws_iam_policy.cmp_policy.arn
}

output "route53_zone_id" {
  description = "The ID of the Route 53 zone"
  value       = aws_route53_zone.private.zone_id
}

output "cmp_instance_profile_arn" {
  description = "The ARN of the instance profile for cmp_service_role"
  value       = aws_iam_instance_profile.cmp_instance_profile.arn
}

output "access_message" {
  value = "Please wait for the service to initialize, about 1 min. Once ready, you can access the service at http://${aws_eip.web_ip.public_ip}:8080"
}