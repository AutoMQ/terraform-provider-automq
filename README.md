# Terraform Provider AutoMQ

Manage AutoMQ BYOC environments and Kafka resources (instances, topics, users, ACLs, mirroring) with Terraform. 

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Getting Started

1. Install the provider plugin via the Terraform Registry (`automq/automq`).
2. Collect your AutoMQ BYOC endpoint plus Service Account Access Key credentials from the AutoMQ console.
3. Configure the provider:

```terraform
terraform {
  required_providers {
    automq = {
      source  = "automq/automq"
      version = "~> 0.1"
    }
  }
}

provider "automq" {
  automq_byoc_endpoint      = var.automq_byoc_endpoint
  automq_byoc_access_key_id = var.automq_byoc_access_key_id
  automq_byoc_secret_key    = var.automq_byoc_secret_key
}
```

> Tip: set `AUTOMQ_BYOC_ENDPOINT`, `AUTOMQ_BYOC_ACCESS_KEY_ID`, and `AUTOMQ_BYOC_SECRET_KEY` environment variables to avoid hard-coding credentials.

## Examples

Reusable configuration samples live under `examples/`. Highlights:
- `examples/provider` – minimal provider configuration.
- `examples/quick-start/aws` – end-to-end instance provisioning, with TLS automation for BYOC on AWS.
- `examples/resources/*` – focused snippets that align with each resource schema (instance, link, mirror group/topic, ACL, etc.).

Use these as references or copy them into your own configuration, updating IDs/secrets as needed.

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
