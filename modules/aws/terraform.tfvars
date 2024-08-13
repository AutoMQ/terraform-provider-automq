
automq_byoc_env_id = "example"
cloud_provider_region         = "ap-southeast-1"
automq_byoc_ec2_instance_type = "m5d.large"

# ami  Defaults to latest image
automq_byoc_env_version = "latest"

# Source of ami: true by aws marketplace, false by the parameter automq_byoc_ami_id.
specified_ami_by_marketplace  = false
automq_byoc_env_console_ami = "ami-067ff8136e9b22196"

# VPC create switch
create_new_vpc = true

# network  Select an existing vpc and the subnet where you need to deploy the application. The subnet should use a public subnet.
automq_byoc_vpc_id = "vpc-022xxxx54103b"
automq_byoc_env_console_public_subnet_id = "subnet-09500xxxxxb6fd28"

# bucket name  If the switch below is set to true, the settings here will be invalid.
automq_byoc_data_bucket_name = "data-bucket"
automq_byoc_ops_bucket_name = "ops-bucket"

# Specify whether a bucket needs to be created by specifying true or false. If you do not fill in the name, automq-data and automq-ops will be created by default.
create_automq_byoc_data_bucket = true
create_automq_byoc_ops_bucket  = true
specific_data_bucket_name      = ""
specific_ops_bucket_name       = ""