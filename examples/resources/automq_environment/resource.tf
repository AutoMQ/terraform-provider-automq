resource "automq_environment" "example" {
  name           = "production"
  description    = "Production environment"
  cloud_provider = "aws"
  region         = "us-east-1"
  scope          = "123456789012"
}
