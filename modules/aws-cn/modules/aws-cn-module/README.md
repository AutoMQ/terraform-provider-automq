## Quickly start the cmp service through terraform

### Environmental preparation

1. Create a VPC that meets the conditions 
Reference documentation:[VPC Create](https://docs.automq.com/zh/automq-cloud/getting-started/create-byoc-environment/aws/step-1-installing-env-with-ami)
You need to get the VPC_ID

2. Parameter configuration

You need to modify the contents of the file `terraform.tfvars` to ensure correct deployment of cmp. The parameters you need to modify are:
```bash
aws_region      = "cn-northwest-1"
aws_access_key  = "AKIAXXXXXXR4SC45"                     # AWS Access Key
aws_secret_key  = "7smvBw1XXXXGMPXxx9FI/LXSnXXXX7W"      # AWS Secret Key
aws_vpc_id      = "vpc-0XXXXc043ae"                      # Fill in the VPC ID created in the previous step.
aws_ami_id      = "ami-035193f2cdb529fda"                # Default is OK
```
> The version of cmp ami can default to the provided version.

### terraform deployment

Execute the command in the directory `/terraform-provider-automq/modules/aws-cn-module`:

```bash
terraform init

terraform plan

terraform apply -auto-approve
```

After successful deployment, some prompt information will be output, such as:

Please wait for the service to initialize, about 1 min. Once ready, you can access the service at http://${aws_eip.web_ip.public_ip}:8080

### cmp initialization

Here you can refer to the official website documentation to complete the initialization:[Init CMP](https://docs.automq.com/zh/automq-cloud/getting-started/create-byoc-environment/aws/step-2-initializing-the-environment)

### Release resources

Execute the command in the directory `/terraform-provider-automq/modules/aws-cn-module`:
```bash
terraform destroy -auto-approve
```