# service name Specifies the service name. If changed, another set of services will be started. If not filled in, the default is automq_byoc.
# Make sure the environment name you enter contains only alphanumeric underscores
automq_byoc_env_name = "example"

# aws info
aws_region                    = "cn-northwest-1"
automq_byoc_ec2_instance_type = "c5.xlarge"  # Can be specified, But you need to ensure that 2 cores are 8g or above

# ami  Defaults to latest cmp image
automq_byoc_version = "latest"

# VPC config
create_vpc = true     # true tf automatically created, false user input available VPC

# network  Select an existing vpc and the subnet where you need to deploy the application. The subnet should use a public subnet.
automq_byoc_vpc_id = "vpc-0ba8fc6b18222d08a"
public_subnet_id = "subnet-0de9f673e8ca59e74"

# bucket name  If the switch below is set to true, the settings here will be invalid.
data_bucket_name = "data-bucket"
ops_bucket_name = "ops-bucket"

# Specify whether a bucket needs to be created by specifying true or false. If you do not fill in the name, automq-data and automq-ops will be created by default.
create_data_bucket       = true
create_ops_bucket        = true
specific_data_bucket_name = ""
specific_ops_bucket_name = ""