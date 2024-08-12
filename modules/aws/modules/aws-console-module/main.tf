# main.tf

provider "aws" {
  region = var.cloud_provider_region
}

data "aws_vpc" "selected" {
  id = var.automq_byoc_vpc_id
}

# Splicing AMI names
locals {
  ami_name_pattern = "automq-control-center-*_linux_amd64"
}

# Get the latest AMI ID
data "aws_ami" "latest_ami" {
  most_recent = true

  filter {
    name = "name"
    values = [local.ami_name_pattern]
  }

  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name = "architecture"
    values = ["x86_64"]
  }

  filter {
    name = "root-device-type"
    values = ["ebs"]
  }

  owners = [327633403396]
}

data "aws_ami" "specific_version_ami" {
  count = var.automq_byoc_env_version == "latest" ? 0 : 1

  most_recent = true

  filter {
    name = "name"
    values = ["automq-control-center-${var.automq_byoc_env_version}_linux_amd64"]
  }

  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name = "architecture"
    values = ["x86_64"]
  }

  filter {
    name = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name = "tag:version"
    values = [var.automq_byoc_env_version]
  }

  owners = [327633403396]
}

resource "aws_security_group" "allow_all" {
  vpc_id = data.aws_vpc.selected.id

  # 上线前只保留8080.并且需要指定CIDR，如果不指定，默认0000
  ingress {
    from_port = 8080
    to_port   = 8080
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 9090
    to_port   = 9090
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 9092
    to_port   = 9092
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 9102
    to_port   = 9102
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 9093
    to_port   = 9093
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 9103
    to_port   = 9103
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Create an IAM role
resource "aws_iam_role" "cmp_role" {
  name = "automq-byoc-service-role-${var.automq_byoc_env_name}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

# Create an IAM policy
resource "aws_iam_policy" "cmp_policy" {
  name        = "automq-byoc-service-policy-${var.automq_byoc_env_name}"
  description = "Custom policy for CMP service"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "iam:CreateServiceLinkedRole"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "iam:AWSServiceName" = "autoscaling.amazonaws.com"
          }
        }
      },
      {
        Sid    = "EC2InstanceProfileManagement"
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = "*"
        Condition = {
          StringLike = {
            "iam:PassedToService" = "ec2.amazonaws.com*"
          }
        }
      },
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameters",
          "pricing:GetProducts",
          "ec2:DescribeImages",
          "ec2:CreateLaunchTemplate",
          "ec2:RebootInstances",
          "ec2:RunInstances",
          "ec2:StopInstances",
          "ec2:TerminateInstances",
          "ec2:CreateKeyPair",
          "ec2:CreateTags",
          "ec2:AttachVolume",
          "ec2:DetachVolume",
          "ec2:DescribeInstances",
          "ec2:DescribeLaunchTemplates",
          "ec2:DescribeLaunchTemplateVersions",
          "ec2:DescribeVolumes",
          "ec2:DescribeSubnets",
          "ec2:DescribeKeyPairs",
          "ec2:DescribeVpcs",
          "ec2:DescribeTags",
          "ec2:DeleteKeyPair",
          "ec2:CreateVolume",
          "ec2:DeleteVolume",
          "ec2:DeleteLaunchTemplate",
          "ec2:DescribeInstanceTypeOfferings",
          "autoscaling:CreateAutoScalingGroup",
          "autoscaling:DescribeAutoScalingGroups",
          "autoscaling:UpdateAutoScalingGroup",
          "autoscaling:DeleteAutoScalingGroup",
          "autoscaling:AttachInstances",
          "autoscaling:DetachInstances",
          "autoscaling:ResumeProcesses",
          "autoscaling:SuspendProcesses",
          "route53:CreateHostedZone",
          "route53:GetHostedZone",
          "route53:ChangeResourceRecordSets",
          "route53:ListHostedZonesByName",
          "route53:ListResourceRecordSets",
          "route53:DeleteHostedZone"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetLifecycleConfiguration",
          "s3:PutLifecycleConfiguration",
          "s3:ListBucket"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetLifecycleConfiguration",
          "s3:PutLifecycleConfiguration",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws-cn:s3:::${var.automq_byoc_data_bucket_name}",
          "arn:aws-cn:s3:::${var.automq_byoc_ops_bucket_name}"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:AbortMultipartUpload",
          "s3:PutObjectTagging",
          "s3:DeleteObject"
        ]
        Resource = [
          "arn:aws-cn:s3:::${var.automq_byoc_data_bucket_name}/*",
          "arn:aws-cn:s3:::${var.automq_byoc_ops_bucket_name}/*"
        ]
      }
    ]
  })
}

# Attach strategies to roles
resource "aws_iam_role_policy_attachment" "cmp_role_attachment" {
  role       = aws_iam_role.cmp_role.name
  policy_arn = aws_iam_policy.cmp_policy.arn
}

# Create an instance profile and bind a role
resource "aws_iam_instance_profile" "cmp_instance_profile" {
  name = "automq-byoc-instance-profile-${var.automq_byoc_env_name}"
  role = aws_iam_role.cmp_role.name
}

# Create an EC2 instance and bind an instance profile
resource "aws_instance" "web" {
  ami           = var.automq_byoc_env_version == "latest" ? data.aws_ami.latest_ami.id : data.aws_ami.specific_version_ami[0].id
  instance_type = var.automq_byoc_ec2_instance_type
  subnet_id     = var.automq_byoc_env_console_public_subnet_id
  vpc_security_group_ids = [aws_security_group.allow_all.id]

  iam_instance_profile = aws_iam_instance_profile.cmp_instance_profile.name

  root_block_device {
    volume_size = 20
    volume_type = "gp2"
  }

  ebs_block_device {
    device_name = "/dev/sdh"
    volume_size = 20
    volume_type = "gp3"
  }

  tags = {
    Name = "automq-byoc-ec2-${var.automq_byoc_env_name}"
  }

  user_data = <<-EOF
              #cloud-config
              bootcmd:
                - |
                  if [ ! -f "/home/admin/config.properties" ]; then
                    touch /home/admin/config.properties
                    echo "cmp.provider.credential=vm-role://${local.aws_iam_instance_profile_arn_encoded}@aws" >> /home/admin/config.properties
                    echo 'cmp.provider.databucket=${var.automq_byoc_data_bucket_name}' >> /home/admin/config.properties
                    echo 'cmp.provider.opsBucket=${var.automq_byoc_ops_bucket_name}' >> /home/admin/config.properties
                    echo 'cmp.provider.instanceSecurityGroup=${aws_security_group.allow_all.id}' >> /home/admin/config.properties
                    echo 'cmp.provider.instanceDNS=${aws_route53_zone.private.zone_id}' >> /home/admin/config.properties
                    echo 'cmp.provider.instanceProfile=${aws_iam_instance_profile.cmp_instance_profile.arn}' >> /home/admin/config.properties
                    echo 'cmp.environmentid=${var.automq_byoc_env_name}' >> /home/admin/config.properties
                  fi
              EOF
}

# Create a Route53 private zone and bind it to the current VPC
resource "aws_route53_zone" "private" {
  name = "${var.automq_byoc_env_name}.automq.private"

  vpc {
    vpc_id = var.automq_byoc_vpc_id
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_eip" "web_ip" {
  instance = aws_instance.web.id
}

# URL encoding instance_profile
locals {
  arn_step1 = replace(aws_iam_instance_profile.cmp_instance_profile.arn, ":", "%3A")
  arn_step2 = replace(local.arn_step1, "/", "%2F")
  arn_step3 = replace(local.arn_step2, "?", "%3F")
  arn_step4 = replace(local.arn_step3, "#", "%23")
  arn_step5 = replace(local.arn_step4, "[", "%5B")
  arn_step6 = replace(local.arn_step5, "]", "%5D")
  arn_step7 = replace(local.arn_step6, "@", "%40")
  arn_step8 = replace(local.arn_step7, "!", "%21")
  arn_step9 = replace(local.arn_step8, "$", "%24")
  arn_step10 = replace(local.arn_step9, "&", "%26")
  arn_step11 = replace(local.arn_step10, "'", "%27")
  arn_step12 = replace(local.arn_step11, "(", "%28")
  arn_step13 = replace(local.arn_step12, ")", "%29")
  arn_step14 = replace(local.arn_step13, "*", "%2A")
  arn_step15 = replace(local.arn_step14, "+", "%2B")
  arn_step16 = replace(local.arn_step15, ",", "%2C")
  arn_step17 = replace(local.arn_step16, ";", "%3B")
  arn_step18 = replace(local.arn_step17, "=", "%3D")
  arn_step19 = replace(local.arn_step18, "%", "%25")
  arn_step20 = replace(local.arn_step19, " ", "%20")
  arn_step21 = replace(local.arn_step20, "<", "%3C")
  arn_step22 = replace(local.arn_step21, ">", "%3E")
  arn_step23 = replace(local.arn_step22, "{", "%7B")
  arn_step24 = replace(local.arn_step23, "}", "%7D")
  arn_step25 = replace(local.arn_step24, "|", "%7C")
  arn_step26 = replace(local.arn_step25, "\\", "%5C")
  arn_step27 = replace(local.arn_step26, "^", "%5E")
  arn_step28 = replace(local.arn_step27, "~", "%7E")

  aws_iam_instance_profile_arn_encoded = local.arn_step28
}