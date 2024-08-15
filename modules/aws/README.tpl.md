# AWS AutoMQ BYOC Environment Terrafrom module

This module is designed for deploying the AutoMQ BYOC (Bring Your Own Cloud) environment using the AWS Provider within an AWS cloud environment.

Upon completion of the installation, the module will output the endpoint of the AutoMQ BYOC environment along with the initial username and password. Users can manage the resources within the environment through the following two methods:

- **Using the Web UI to manage resources**: This method allows users to manage instances, topics, ACLs, and other resources through a web-based interface.
- **Using Terraform to manage resources**: This method requires users to access the AutoMQ BYOC environment via a web browser for the first time to create a Service Account. Subsequently, users can manage resources within the environment using the Service Account's Access Key and the AutoMQ Terraform Provider.

For managing instances, topics, and other resources within the AutoMQ BYOC environment using the AutoMQ Terraform Provider, please refer to the [documentation](https://docs.automq.com/automq-cloud/manage-identities-and-access).

# Module Usage
Use this module to install the AutoMQ BYOC environment, supporting two modes:

- **Create a new VPC and install**: Recommended only for POC or other testing scenarios. In this mode, the user only needs to specify the region, and resources including VPC, Endpoint, Security Group, S3 Bucket, etc., will be created. After testing, all resources can be destroyed with one click.
- **Install using an existing VPC**: Recommended for production environments. In this mode, the user needs to provide a VPC, subnet, and S3 Bucket that meet the requirements. AutoMQ will deploy the BYOC environment console to the user-specified subnet.

## Create a new VPC and install

```terraform
module "automq_byoc" {
  source = "terraform-automq-modules/automq-byoc-console"

  # Set the identifier for the environment to be installed. This ID will be used for naming internal resources. The environment ID supports only uppercase and lowercase English letters, numbers, and hyphens (-). It must start with a letter and is limited to a length of 32 characters.
  automq_byoc_env_id                       = "example" 

  # Set the target regionId of aws
  cloud_provider_region                    = "ap-southeast-1"  
}
```

## Install using an existing VPC

To install the AutoMQ BYOC environment using an existing VPC, ensure your existing VPC meets the necessary requirements. You can find the detailed requirements in the [doc](https://docs.automq.com/zh/automq-cloud/getting-started/create-byoc-environment/aws/step-1-installing-env-with-ami#%E6%AD%A5%E9%AA%A4-3%E5%90%AF%E5%8A%A8-ec2-%E5%AE%9E%E4%BE%8B%E5%AE%89%E8%A3%85%E7%8E%AF%E5%A2%83).

```terraform
module "automq_byoc" {
  source = "terraform-automq-modules/automq-byoc-console"
  
  # Set the identifier for the environment to be installed. This ID will be used for naming internal resources. The environment ID supports only uppercase and lowercase English letters, numbers, and hyphens (-). It must start with a letter and is limited to a length of 32 characters.  
  automq_byoc_env_id                       = "example"

  # Set the target regionId of aws    
  cloud_provider_region                    = "ap-southeast-1" 

  # Set this switch to false, use existed vpc  
  create_new_vpc                           = false   

  # Set this existed vpc
  automq_byoc_vpc_id                       = "vpc-022xxxx54103b"  

  # Set the subnet for deploying the AutoMQ environment console. This subnet must support internet access, and EC2 instances created within this subnet must be able to access the internet.
  automq_byoc_env_console_public_subnet_id = "subnet-09500xxxxxb6fd28"  
  
}
```