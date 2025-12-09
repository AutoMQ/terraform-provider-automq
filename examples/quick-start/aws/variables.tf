variable "automq_byoc_endpoint" {
  description = "AutoMQ control plane endpoint"
  type        = string
}

variable "automq_byoc_access_key_id" {
  description = "Access key for AutoMQ BYOC"
  type        = string
}

variable "automq_byoc_secret_key" {
  description = "Secret key for AutoMQ BYOC"
  type        = string
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region where the EKS cluster is located"
  type        = string
  default     = "ap-northeast-1"
}

variable "environment_id" {
  description = "Target AutoMQ environment ID"
  type        = string
}

variable "instance_name" {
  description = "Kafka instance name"
  type        = string
}

variable "description" {
  description = "Kafka instance description"
  type        = string
  default     = ""
}

variable "automq_version" {
  description = "Kafka instance version"
  type        = string
  default     = "5.2.0"
}

variable "reserved_aku" {
  description = "Reserved AKU size"
  type        = number
}

variable "deploy_type" {
  description = "Deployment platform (IAAS or K8S)"
  type        = string
  default     = "IAAS"
}

variable "networks" {
  description = "Network configuration specifying availability zones"
  type = list(object({
    zone    = string
    subnets = optional(list(string))
  }))
}

variable "dns_zone" {
  description = "Route53 hosted zone ID"
  type        = string
  default     = null
}

variable "data_buckets" {
  description = "Object storage bucket configuration"
  type = list(object({
    bucket_name = string
  }))
  default = []
}

variable "instance_role" {
  description = "IAM role ARN"
  type        = string
  default     = null
}

variable "kubernetes_cluster_id" {
  description = "Kubernetes cluster identifier"
  type        = string
}

variable "kubernetes_node_groups" {
  description = "Kubernetes node group identifiers"
  type = list(object({
    id = string
  }))
}

variable "kubernetes_namespace" {
  description = "Kubernetes namespace"
  type        = string
  default     = null
}

variable "kubernetes_service_account" {
  description = "Kubernetes service account"
  type        = string
  default     = null
}

variable "wal_mode" {
  description = "WAL mode (EBSWAL or S3WAL)"
  type        = string
  default     = "EBSWAL"
}

variable "instance_configs" {
  description = "Additional broker configurations"
  type        = map(string)
  default     = {}
}

variable "metrics_exporter" {
  description = "Prometheus exporter configuration"
  type = object({
    prometheus = optional(object({
      enabled        = optional(bool)
      auth_type      = optional(string)
      endpoint       = optional(string)
      prometheus_arn = optional(string)
      username       = optional(string)
      password       = optional(string)
      token          = optional(string)
      labels         = optional(map(string))
    }))
  })
  default = null
}

variable "table_topic" {
  description = "Table topic integration configuration"
  type = object({
    warehouse          = string
    catalog_type       = string
    metastore_uri      = optional(string)
    hive_auth_mode     = optional(string)
    kerberos_principal = optional(string)
    user_principal     = optional(string)
    keytab_file        = optional(string)
    krb5conf_file      = optional(string)
  })
  default = null
}

variable "security_authentication_methods" {
  description = "Authentication methods"
  type        = list(string)
  default     = ["sasl"]
}

variable "security_transit_encryption_modes" {
  description = "Transit encryption modes"
  type        = list(string)
  default     = ["tls"]
}

variable "security_data_encryption_mode" {
  description = "Data encryption mode"
  type        = string
  default     = "CPMK"
}
