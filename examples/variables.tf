variable "api_key" {
  type        = string
  description = "Crunchy Bridge API key - Application ID"
}

variable "api_secret" {
  type        = string
  description = "Crunchy Bridge API key - Application Secret"
  sensitive   = true
}

variable "example_id" {
  type        = string
  description = "Cluster ID to be used in example .tf files"
}
