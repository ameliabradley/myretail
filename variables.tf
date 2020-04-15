variable "project_id" {
    type = string
}

variable "datastore_id" {
    type = string
}

variable "region" {
    type = string
}

variable "bucket_name" {
    type = string
}

variable "bucket_force_destroy" {
  type        = bool
  default     = false
  description = "When deleting the GCS bucket containing the cloud function, delete all objects in the bucket first."
}