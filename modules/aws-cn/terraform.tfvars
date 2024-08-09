# service name Specifies the service name. If changed, another set of services will be started. If not filled in, the default is cmp_service.
service_name = "cmp-service"

# aws info
aws_region = "cn-northwest-1"

# ami  Defaults to 1.1.3 cmp image
aws_ami_id = "ami-035193f2cdb529fda"

# network  Select an existing vpc and the subnet where you need to deploy the application. The subnet should use a public subnet.
aws_vpc_id = "vpc-0ba8fc6b18222d08a"
subnet_id = "subnet-0de9f673e8ca59e74"

# bucket name  If the switch below is set to true, the settings here will be invalid.
data_bucket_name = "data-bucket"
ops_bucket_name = "ops-bucket"

# Specify whether a bucket needs to be created by specifying true or false. If you do not fill in the name, automq-data and automq-ops will be created by default.
create_data_bucket        = true
create_ops_bucket         = true
specific_data_bucket_name = "automq-data"
specific_ops_bucket_name  = "automq-ops"