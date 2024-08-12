
automq_byoc_env_name = "example"
cloud_provider_region         = "cn-northwest-1"
automq_byoc_ec2_instance_type = "c5.xlarge"

# ami  Defaults to latest cmp image
automq_byoc_env_version = "latest"

# VPC create switch
create_new_vpc = true

# network  Select an existing vpc and the subnet where you need to deploy the application. The subnet should use a public subnet.
automq_byoc_vpc_id = "vpc-0bxxxxx22d08a"
automq_byoc_env_console_public_subnet_id = "subnet-0de9xxxxxx59e74"

# bucket name  If the switch below is set to true, the settings here will be invalid.
automq_byoc_data_bucket_name = "data-bucket"
automq_byoc_ops_bucket_name = "ops-bucket"

# Specify whether a bucket needs to be created by specifying true or false. If you do not fill in the name, automq-data and automq-ops will be created by default.
create_automq_byoc_data_bucket = true
create_automq_byoc_ops_bucket  = true
specific_data_bucket_name      = ""
specific_ops_bucket_name       = ""