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

<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 5.62.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_automq_byoc"></a> [automq\_byoc](#module\_automq\_byoc) | ./modules/aws-console-module | 1.0.0   |
| <a name="module_automq_byoc_data_bucket_name"></a> [automq\_byoc\_data\_bucket\_name](#module\_automq\_byoc\_data\_bucket\_name) | terraform-aws-modules/s3-bucket/aws | 4.1.2   |
| <a name="module_automq_byoc_ops_bucket_name"></a> [automq\_byoc\_ops\_bucket\_name](#module\_automq\_byoc\_ops\_bucket\_name) | terraform-aws-modules/s3-bucket/aws | 4.1.2   |
| <a name="module_automq_byoc_vpc"></a> [automq\_byoc\_vpc](#module\_automq\_byoc\_vpc) | terraform-aws-modules/vpc/aws | 5.0.0   |

## Resources

| Name | Type |
|------|------|
| [aws_eip.nat](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eip) | resource |
| [aws_security_group.endpoint_sg](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_vpc_endpoint.ec2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_endpoint) | resource |
| [aws_vpc_endpoint.s3](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_endpoint) | resource |
| [aws_availability_zones.available](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/availability_zones) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_automq_byoc_data_bucket_name"></a> [automq\_byoc\_data\_bucket\_name](#input\_automq\_byoc\_data\_bucket\_name) | Set the existed object storage bucket for that used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket. | `string` | n/a |    no    |
| <a name="input_automq_byoc_ec2_instance_type"></a> [automq\_byoc\_ec2\_instance\_type](#input\_automq\_byoc\_ec2\_instance\_type) | Can be specified, But you need to ensure that 2 cores are 8g or above | `string` | `"m5d.large"` |    no    |
| <a name="input_automq_byoc_env_console_ami"></a> [automq\_byoc\_env\_console\_ami](#input\_automq\_byoc\_env\_console\_ami) | When obtaining ami id from non-cloud market, manually specify ami id. | `string` | n/a |    no    |
| <a name="input_automq_byoc_env_console_public_subnet_id"></a> [automq\_byoc\_env\_console\_public\_subnet\_id](#input\_automq\_byoc\_env\_console\_public\_subnet\_id) | Select a subnet for deploying the AutoMQ BYOC environment console. Ensure that the chosen subnet supports public access. | `string` | n/a |    no    |
| <a name="input_automq_byoc_env_id"></a> [automq\_byoc\_env\_id](#input\_automq\_byoc\_env\_id) | This parameter is used to create resources within the environment. Additionally, all cloud resource names will incorporate this parameter as part of their names.This parameter supports only numbers, uppercase and lowercase English letters, and hyphens. It must start with a letter and is limited to a length of 32 characters. | `string` | n/a |   yes    |
| <a name="input_automq_byoc_env_version"></a> [automq\_byoc\_env\_version](#input\_automq\_byoc\_env\_version) | Set the version for the AutoMQ BYOC environment console. It is recommended to keep the default value, which is the latest version. | `string` | n/a |    no    |
| <a name="input_automq_byoc_ops_bucket_name"></a> [automq\_byoc\_ops\_bucket\_name](#input\_automq\_byoc\_ops\_bucket\_name) | Set the existed object storage bucket for that used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket. | `string` | n/a |    no    |
| <a name="input_automq_byoc_vpc_id"></a> [automq\_byoc\_vpc\_id](#input\_automq\_byoc\_vpc\_id) | The ID of the VPC | `string` | n/a |    no    |
| <a name="input_cloud_provider_region"></a> [cloud\_provider\_region](#input\_cloud\_provider\_region) | Set the cloud provider's region. AutoMQ will deploy to this region. | `string` | n/a |   yes    |
| <a name="input_create_automq_byoc_data_bucket"></a> [create\_automq\_byoc\_data\_bucket](#input\_create\_automq\_byoc\_data\_bucket) | This parameter controls whether to create a new bucket. If it is a POC scenario or there is no available bucket, set it to true. If there is already a suitable bucket, set it to false. | `bool` | `true` |    no    |
| <a name="input_create_automq_byoc_ops_bucket"></a> [create\_automq\_byoc\_ops\_bucket](#input\_create\_automq\_byoc\_ops\_bucket) | This parameter controls whether to create a new bucket. If it is a POC scenario or there is no available bucket, set it to true. If there is already a suitable bucket, set it to false. | `bool` | `true` |    no    |
| <a name="input_create_new_vpc"></a> [create\_new\_vpc](#input\_create\_new\_vpc) | This setting determines whether to create a new VPC. If set to true, a new VPC spanning three availability zones will be automatically created, which is recommended only for POC scenarios. For production scenario using AutoMQ, you should provide the VPC where the current Kafka application resides and check the current VPC against the requirements specified in the documentation. | `bool` | `true` |    no    |
| <a name="input_specific_data_bucket_name"></a> [specific\_data\_bucket\_name](#input\_specific\_data\_bucket\_name) | Set up a new object storage bucket that will be used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket. | `string` | `"automq-data"` |    no    |
| <a name="input_specific_ops_bucket_name"></a> [specific\_ops\_bucket\_name](#input\_specific\_ops\_bucket\_name) | Set up a new object storage bucket that will be used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket. | `string` | `"automq-ops"` |    no    |
| <a name="input_specified_ami_by_marketplace"></a> [specified\_ami\_by\_marketplace](#input\_specified\_ami\_by\_marketplace) | Specifies the switch to obtain ami id from the cloud market. If true, it means to obtain the specified version of ami id through the cloud market. Otherwise, it needs to be specified manually. | `bool` | `true` |    no    |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_automq_byoc_data_bucket_arn"></a> [automq\_byoc\_data\_bucket\_arn](#output\_automq\_byoc\_data\_bucket\_arn) | Data storage bucket arn. |
| <a name="output_automq_byoc_data_bucket_name"></a> [automq\_byoc\_data\_bucket\_name](#output\_automq\_byoc\_data\_bucket\_name) | The object storage bucket for that used to store message data generated by applications. The message data Bucket must be separate from the Ops Bucket. |
| <a name="output_automq_byoc_env_console_ami"></a> [automq\_byoc\_env\_console\_ami](#output\_automq\_byoc\_env\_console\_ami) | Mirror ami id of AutoMQ BYOC Console. |
| <a name="output_automq_byoc_env_console_ec2_instance_ip"></a> [automq\_byoc\_env\_console\_ec2\_instance\_ip](#output\_automq\_byoc\_env\_console\_ec2\_instance\_ip) | The instance IP of the deployed AutoMQ BYOC control panel. You can access the service through this IP. |
| <a name="output_automq_byoc_env_console_public_subnet_id"></a> [automq\_byoc\_env\_console\_public\_subnet\_id](#output\_automq\_byoc\_env\_console\_public\_subnet\_id) | AutoMQ WebUI is deployed under this subnet. |
| <a name="output_automq_byoc_env_id"></a> [automq\_byoc\_env\_id](#output\_automq\_byoc\_env\_id) | This parameter is used to create resources within the environment. Additionally, all cloud resource names will incorporate this parameter as part of their names.This parameter supports only numbers, uppercase and lowercase English letters, and hyphens. It must start with a letter and is limited to a length of 32 characters. |
| <a name="output_automq_byoc_env_webui_address"></a> [automq\_byoc\_env\_webui\_address](#output\_automq\_byoc\_env\_webui\_address) | Address accessed by AutoMQ BYOC service |
| <a name="output_automq_byoc_instance_id"></a> [automq\_byoc\_instance\_id](#output\_automq\_byoc\_instance\_id) | AutoMQ BYOC Console instance ID. |
| <a name="output_automq_byoc_instance_profile_arn"></a> [automq\_byoc\_instance\_profile\_arn](#output\_automq\_byoc\_instance\_profile\_arn) | Instance configuration file ARN |
| <a name="output_automq_byoc_ops_bucket_arn"></a> [automq\_byoc\_ops\_bucket\_arn](#output\_automq\_byoc\_ops\_bucket\_arn) | Ops storage bucket arn. |
| <a name="output_automq_byoc_ops_bucket_name"></a> [automq\_byoc\_ops\_bucket\_name](#output\_automq\_byoc\_ops\_bucket\_name) | The object storage bucket for that used to store AutoMQ system logs and metrics data for system monitoring and alerts. This Bucket does not contain any application business data. The Ops Bucket must be separate from the message data Bucket. |
| <a name="output_automq_byoc_policy_arn"></a> [automq\_byoc\_policy\_arn](#output\_automq\_byoc\_policy\_arn) | AutoMQ BYOC is bound to a custom policy on the role arn. |
| <a name="output_automq_byoc_role_arn"></a> [automq\_byoc\_role\_arn](#output\_automq\_byoc\_role\_arn) | AutoMQ BYOC is bound to the role arn of the Console. |
| <a name="output_automq_byoc_security_group_name"></a> [automq\_byoc\_security\_group\_name](#output\_automq\_byoc\_security\_group\_name) | Security group bound to the AutoMQ BYOC service. |
| <a name="output_automq_byoc_vpc_id"></a> [automq\_byoc\_vpc\_id](#output\_automq\_byoc\_vpc\_id) | AutoMQ BYOC is deployed in this VPC. |
| <a name="output_automq_byoc_vpc_route53_zone_id"></a> [automq\_byoc\_vpc\_route53\_zone\_id](#output\_automq\_byoc\_vpc\_route53\_zone\_id) | Route53 bound to the VPC. |
<!-- END_TF_DOCS -->