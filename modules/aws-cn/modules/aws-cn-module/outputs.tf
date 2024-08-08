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

output "access_message" {
  value = "Please wait for the service to initialize, about 1 min. Once ready, you can access the service at http://${aws_eip.web_ip.public_ip}:8080"
}