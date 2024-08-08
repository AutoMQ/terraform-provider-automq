# main.tf

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

data "aws_vpc" "selected" {
  id = var.aws_vpc_id
}

resource "aws_security_group" "allow_all" {
  vpc_id = data.aws_vpc.selected.id

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # 新增端口 9090
  ingress {
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # 新增端口 9092
  ingress {
    from_port   = 9092
    to_port     = 9092
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # 新增端口 9102
  ingress {
    from_port   = 9102
    to_port     = 9102
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # 新增端口 9093
  ingress {
    from_port   = 9093
    to_port     = 9093
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # 新增端口 9103
  ingress {
    from_port   = 9103
    to_port     = 9103
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# 创建 IAM 角色
resource "aws_iam_role" "cmp_role" {
  name = "cmp_service_role"

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

# 创建 IAM 策略
resource "aws_iam_policy" "cmp_policy" {
  name        = "cmp_service_policy"
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
        Sid = "EC2InstanceProfileManagement"
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
          "arn:aws-cn:s3:::${var.data_bucket_name}",
          "arn:aws-cn:s3:::${var.ops_bucket_name}"
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
          "arn:aws-cn:s3:::${var.data_bucket_name}/*",
          "arn:aws-cn:s3:::${var.ops_bucket_name}/*"
        ]
      }
    ]
  })
}

# 附加策略到角色
resource "aws_iam_role_policy_attachment" "cmp_role_attachment" {
  role       = aws_iam_role.cmp_role.name
  policy_arn = aws_iam_policy.cmp_policy.arn
}

# 创建实例配置文件并绑定角色
resource "aws_iam_instance_profile" "cmp_instance_profile" {
  name = "cmp_instance_profile"
  role = aws_iam_role.cmp_role.name
}

# 创建 EC2 实例并绑定实例配置文件
resource "aws_instance" "web" {
  ami                    = var.aws_ami_id
  instance_type          = "c5.xlarge"
  subnet_id              = var.subnet_id
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
    Name = "cmp-service"
  }

  user_data = <<-EOF
              #cloud-config
              bootcmd:
                - |
                  if [ ! -f "/home/admin/config.properties" ]; then
                    touch /home/admin/config.properties
                    echo "cmp.provider.credential=vm-role://${local.aws_iam_instance_profile_arn_encoded}@aws-cn" >> /home/admin/config.properties
                    echo 'cmp.provider.databucket=${var.data_bucket_name}' >> /home/admin/config.properties
                    echo 'cmp.provider.opsBucket=${var.ops_bucket_name}' >> /home/admin/config.properties
                    echo 'cmp.provider.instanceSecurityGroup=${aws_security_group.allow_all.id}' >> /home/admin/config.properties
                    echo 'cmp.provider.instanceDNS=${aws_route53_zone.private.zone_id}' >> /home/admin/config.properties
                    echo 'cmp.provider.instanceProfile=${aws_iam_instance_profile.cmp_instance_profile.arn}' >> /home/admin/config.properties
                    echo 'cmp.environmentid=example-service-instance-id' >> /home/admin/config.properties
                  fi
              EOF
}

# 创建 Route53 private zone 并绑定到当前 VPC
resource "aws_route53_zone" "private" {
  name = "cmp_route53_zone"

  vpc {
    vpc_id = var.aws_vpc_id
  }
}

resource "aws_eip" "web_ip" {
  instance = aws_instance.web.id
}

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