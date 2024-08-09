## Quickly start the cmp service through terraform

### Environmental preparation

#### Create a VPC that meets the conditions

Reference
documentation:[VPC Create](https://docs.automq.com/zh/automq-cloud/getting-started/create-byoc-environment/aws/step-1-installing-env-with-ami)
You need to get the vpc_id and subnet_id

#### Parameter configuration

You need to prepare an aws user. **Please ensure that the user has the following permissions**:

- AmazonVPCFullAccess: Permissions to manage VPCs.
- AmazonEC2FullAccess: Permission to manage EC2 products.
- AmazonS3FullAccess: Manage permissions for object storage S3.
- AmazonRoute53FullAccess: Permission to manage Route 53 services.

And configure the aws environment:

```bash
PS Path:xxx > aws configure
AWS Access Key ID [****************SC45]:
AWS Secret Access Key [****************PB7W]:
Default region name [cn-northwest-1]:
Default output format [json]:
```

#### Configure network and buckets

Modify the `terraform.tfvars` file in the current module root directory and configure the vpc and public subnet id
created in the above process.
And configure the bucket as needed. For the bucket here, you can choose the s3 bucket you created yourself, or you can
specify the name and use the bucket created by terraform.

### terraform deployment

Execute these commands in the directory `/terraform-provider-automq/modules/aws-cn`:

```bash
terraform init

terraform plan

terraform apply -auto-approve
```

After successful deployment, some prompt information will be output, such as:

**Please wait for the service to initialize, about 1 min. Once ready, you can access the service at http:
//${aws_eip.web_ip.public_ip}:8080**

Later, you can access the successfully deployed cmp service through this address

### cmp initialization

Here you can refer to the official website documentation to complete the
initialization:[Init CMP](https://docs.automq.com/zh/automq-cloud/getting-started/create-byoc-environment/aws/step-2-initializing-the-environment)

### Release resources

Since these operations may cost some money, please release resources when they are not needed to avoid waste.

1. Release the instance of the AutoMQ cluster through the cmp control panel and wait for the deletion to be successful.
2. Release the resources created by terraform

Execute these commands in the directory `/terraform-provider-automq/modules/aws-cn`:
```bash
terraform destroy -auto-approve
```