variable "automq_byoc_endpoint" {
  description = "AutoMQ Control Plane API endpoint"
  type        = string
}

variable "automq_byoc_access_key_id" {
  description = "Access Key ID for an AutoMQ Service Account"
  type        = string
}

variable "automq_byoc_secret_key" {
  description = "Secret Access Key for an AutoMQ Service Account"
  type        = string
  sensitive   = true
}

variable "environment_id" {
  description = "Target GCP BYOC environment ID"
  type        = string
}

variable "instance_name" {
  description = "Kafka instance name"
  type        = string
}

variable "automq_version" {
  description = "AutoMQ instance version supported by the target environment"
  type        = string
}

variable "reserved_aku" {
  description = "Reserved AKU size"
  type        = number
}

variable "zone" {
  description = "GCP zone used by the selected GKE node pool, for example us-central1-a"
  type        = string
}

variable "instance_type" {
  description = "GCE machine type used by the selected GKE node pool, for example n2d-standard-4"
  type        = string
}

variable "load_balancer_subnet_id" {
  description = "GCP subnetwork full resource name: projects/<project>/regions/<region>/subnetworks/<name>"
  type        = string
}

variable "schedule_spec" {
  description = "Kubernetes affinity and tolerations YAML matching the selected GKE node pool"
  type        = string
}
