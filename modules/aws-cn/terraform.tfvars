# aws infos
aws_region     = "cn-northwest-1"
aws_access_key = "AKIAUYSDY7YCI3R4SC45"
aws_secret_key = "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"

# ami   默认为 1.1.3 的 cmp 镜像
aws_ami_id     = "ami-035193f2cdb529fda"

# network    选择已有的 vpc 和你需要部署应用的子网
aws_vpc_id     = "vpc-0ac01fa7ea7c043ae"
subnet_id      = "subnet-00a963d37d8a01963"

# bucket name   如果下方的开关设置为true，这里的设置会失效
data_bucket_name = ""
ops_bucket_name  = ""

# 通过指定 true 或者 false 来指定是否需要创建bucket
create_data_bucket = true
create_ops_bucket  = true